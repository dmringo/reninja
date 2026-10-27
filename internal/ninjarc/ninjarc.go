package ninjarc

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"slices"
	"strings"

	"github.com/buildbuddy-io/reninja/internal/command_line"
	"github.com/buildbuddy-io/reninja/internal/disk"
	"github.com/buildbuddy-io/reninja/internal/project_root"
	"github.com/buildbuddy-io/reninja/internal/util"
)

const (
	workspacePrefix = `%workspace%/`
)

var (
	ignoreConfig = flag.Bool("norc", false, "ignore all RC files (including ones in default locations)")
	ninjaRCFile  = flag.String("ninjarc", "", "path to a ninjarc file to parse")
	configFlag   = flag.String("config", "", "ninjarc configuration to apply")

	rcFileLocations = []string{".ninjarc", "%workspace%/.ninjarc", "~/.ninjarc", "/etc/.ninjarc"}
)

// Package ninjarc defines a simple RC format to configure ninja flags without
// having to specify them on the command line for every invocation.
//
// The format is very simple. Here is an example:
//
// build:local --bes_backend="grpc://localhost:1985"
// build:local --remote_cache="grpc://localhost:1985"
// build:local --results_url="http://localhost:8080"
//
// This config specifies that for "build" commands, when the --config flag value
// is "local", the `--bes_backend`, `--remote_cache` and `--results_url` flags
// will be set.
//
// Here's another example:
//
// clean --bes_backend="grpc://localhost:1985"
// clean --remote_cache="grpc://localhost:1985"
// clean --results_url="http://localhost:8080"
//
// This config specifies that for "clean" commands, regardless of config flag
// value, the `--bes_backend`, `--remote_cache` and `--results_url` flags will
// be set.

// RcRule is a rule parsed from a ninjarc file.
type RcRule struct {
	Phase string

	// Make Config a pointer to a string so we can distinguish default configs
	// from configs with blank names.
	Config *string

	// Tokens contains the raw (non-canonicalized) tokens in the rule.
	Tokens []string
}

// resolvePath resolves a ninjarc path by expanding %workspace%/ and ~
// prefixes, cleaning the path, and making it relative to the filesystem root
// backing the RC FS. fs.FS paths must be unrooted, so drive letters and
// leading separators are stripped. The FS root is the volume containing
// workspaceDir (on Windows) or "/" elsewhere.
func resolvePath(workspaceDir, p string) string {
	if strings.HasPrefix(p, workspacePrefix) {
		p = filepath.Join(workspaceDir, p[len(workspacePrefix):])
	}

	if strings.HasPrefix(p, "~") {
		currentUser, err := user.Current()
		if err == nil {
			p = filepath.Join(currentUser.HomeDir, strings.TrimPrefix(p, "~"))
		}
	}

	p = filepath.Clean(p)
	rootVolume := filepath.VolumeName(workspaceDir)
	if volume := filepath.VolumeName(p); volume != "" {
		if rootVolume != "" && !strings.EqualFold(volume, rootVolume) {
			// The file lives on another volume than the FS root. Leave it
			// absolute so the FS layer rejects it instead of opening a
			// same-named path on the wrong drive.
			return filepath.ToSlash(p)
		}
		p = strings.TrimPrefix(p, volume)
	}

	// fs.FS paths must be unrooted.
	return strings.TrimLeft(filepath.ToSlash(p), "/")
}

// rcFSRoot returns the root directory that all resolved RC paths are
// relative to.
func rcFSRoot(workspaceDir string) string {
	if volume := filepath.VolumeName(workspaceDir); volume != "" {
		return volume + string(filepath.Separator)
	}
	return string(filepath.Separator)
}

// AppendRcRulesFromFile reads and lexes the provided rc file and appends the
// args to the provided configs based on the detected phase and name.
//
// configs is a map keyed by config name where the values are maps keyed by
// phase name where the values are lists containing all the rules for that
// config in the order they are encountered.
func AppendRcRulesFromFile(fsys fs.FS, workspaceDir string, path string, namedConfigs map[string]map[string][]string, defaultConfig map[string][]string, importStack []string, optional bool) error {
	if slices.Contains(importStack, path) {
		return fmt.Errorf("circular import detected: %s -> %s", strings.Join(importStack, " -> "), path)
	}
	importStack = append(importStack, path)
	file, err := fsys.Open(path)
	if err != nil {
		if optional {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Handle line continuations (lines can end with "\" to effectively
		// continue the same line)
		for strings.HasSuffix(line, `\`) && scanner.Scan() {
			line = line[:len(line)-1] + scanner.Text()
		}
		tokens, err := command_line.Split(line)
		if err != nil {
			return fmt.Errorf("Error parsing ninjarc: %s\nFailed to lex '%s'.", err, line)
		}
		if len(tokens) == 0 {
			// blank line
			continue
		}
		if tokens[0] == "import" || tokens[0] == "try-import" {
			isOptional := tokens[0] == "try-import"
			if len(tokens) < 2 {
				return fmt.Errorf("Error parsing ninjarc: missing import path in '%s'", line)
			}
			importPath := tokens[1]
			if err := AppendRcRulesFromFile(fsys, workspaceDir, resolvePath(workspaceDir, importPath), namedConfigs, defaultConfig, importStack, isOptional); err != nil {
				return err
			}
			continue
		}

		rule, err := parseRcRule(tokens)
		if err != nil {
			util.Errorf("Error parsing ninjarc option: %s", err.Error())
			continue
		}
		if rule == nil {
			continue
		}
		if rule.Config == nil {
			defaultConfig[rule.Phase] = append(defaultConfig[rule.Phase], rule.Tokens...)
			continue
		}

		config, ok := namedConfigs[*rule.Config]
		if !ok {
			config = make(map[string][]string)
			namedConfigs[*rule.Config] = config
		}
		config[rule.Phase] = append(config[rule.Phase], rule.Tokens...)
	}
	return scanner.Err()
}

func parseRcRule(tokens []string) (*RcRule, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("unexpected empty line")
	}
	if len(tokens) == 1 {
		// ninja ignores .ninjarc lines consisting of a single shlex token
		return nil, nil
	}
	phase := tokens[0]
	var configName *string
	if colonIndex := strings.Index(tokens[0], ":"); colonIndex != -1 {
		phase = tokens[0][:colonIndex]
		v := tokens[0][colonIndex+1:]
		configName = &v
	}

	return &RcRule{
		Phase:  phase,
		Config: configName,
		Tokens: tokens[1:],
	}, nil
}

type RCConfig struct {
	namedRcRules   map[string]map[string][]string
	defaultRcRules map[string][]string
}

func (c *RCConfig) Apply(toolName string, config string, flagSet *flag.FlagSet, args []string) {
	expandedValues := make([]string, 0)
	seenConfigs := make(map[string]struct{}, 0)
	tools := []string{"common", toolName}

	var expandRules func(config string)
	addOptionToExpanded := func(opt string) {
		key, val, ok := strings.Cut(strings.TrimLeft(opt, "-"), "=")
		// if we encounter a flag like "--config=dev", then
		// expand the rules from "dev".
		if ok && key == "config" {
			expandRules(val)
		}
		expandedValues = append(expandedValues, opt)
	}

	expandRules = func(config string) {
		if _, ok := seenConfigs[config]; ok {
			return
		}
		seenConfigs[config] = struct{}{}
		optsByTool := c.namedRcRules[config]

		for _, toolName := range tools {
			for _, opt := range optsByTool[toolName] {
				addOptionToExpanded(opt)
			}
		}
		delete(seenConfigs, config)
	}

	for _, toolName := range tools {
		for _, opt := range c.defaultRcRules[toolName] {
			addOptionToExpanded(opt)
		}
	}

	expandRules(config)
	expandedValues = append(expandedValues, args...)
	flagSet.Parse(expandedValues)
}

// ParseRCFiles parses the provided rc files in the given workspace into Configs
// and returns a map of the named configs as well as the default (unnamed)
// Config.
func ParseRCFiles(fsys fs.FS, workspaceDir string, filePaths ...string) (*RCConfig, error) {
	seen := make(map[string]struct{}, len(filePaths))
	namedRcRules := map[string]map[string][]string{}
	defaultRcRules := map[string][]string{}

	for _, filePath := range filePaths {
		r := resolvePath(workspaceDir, filePath)
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}

		err := AppendRcRulesFromFile(fsys, workspaceDir, r, namedRcRules, defaultRcRules, nil /*=importStack*/, true)
		if err != nil {
			return nil, err
		}
	}

	return &RCConfig{
		namedRcRules:   namedRcRules,
		defaultRcRules: defaultRcRules,
	}, nil

}

func ParseAndApplyRCFilesToFlags(toolName string) error {
	flag.Parse() // Parse flags once to ensure we have --config, --ninjarc, and --norc

	if *ignoreConfig {
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	workspaceDir := project_root.Root()

	filePaths := make([]string, 0, len(rcFileLocations)+1)
	for _, p := range rcFileLocations {
		// The CWD .ninjarc is relative to the process, not to the RC FS root.
		if p == ".ninjarc" {
			p = filepath.Join(cwd, p)
		}
		filePaths = append(filePaths, p)
	}
	if *ninjaRCFile != "" {
		p := *ninjaRCFile
		if strings.HasPrefix(p, workspacePrefix) {
			p = filepath.Join(workspaceDir, p[len(workspacePrefix):])
		}
		if strings.HasPrefix(p, "~") {
			currentUser, err := user.Current()
			if err == nil {
				p = filepath.Join(currentUser.HomeDir, strings.TrimPrefix(p, "~"))
			}
		}
		if !filepath.IsAbs(p) {
			abs, err := filepath.Abs(p)
			if err != nil {
				return err
			}
			p = abs
		}
		if !disk.FileExists(p) {
			util.Fatalf("ninjarc file '%s' not found", p)
		}
		filePaths = append(filePaths, p)
	}

	rcConfig, err := ParseRCFiles(os.DirFS(rcFSRoot(workspaceDir)), workspaceDir, filePaths...)
	if err != nil {
		return err
	}

	rcConfig.Apply(toolName, *configFlag, flag.CommandLine, os.Args[1:])
	return nil
}
