// Package clparser extracts dependencies from MSVC /showIncludes output.
package clparser

import (
	"path/filepath"
	"strings"

	"github.com/buildbuddy-io/reninja/internal/util"
)

const defaultDepsPrefix = "Note: including file: "

// IsSystemInclude reports whether an include path is part of the MSVC or
// Windows SDK installation. Ninja excludes these from the deps log.
func IsSystemInclude(path string) bool {
	lower := strings.ToLower(path)
	return strings.Contains(lower, "program files") ||
		strings.Contains(lower, "microsoft visual studio")
}

// NormalizeInclude converts an include path reported by cl.exe to the
// canonical form stored in .ninja_deps. Paths on the same drive as cwd are
// made relative to it; paths on other drives are kept absolute.
func NormalizeInclude(include, cwd string) (string, error) {
	p := include
	if !filepath.IsAbs(p) {
		p = filepath.Join(cwd, p)
	}
	p = filepath.Clean(p)

	if rel, err := filepath.Rel(cwd, p); err == nil {
		p = rel
	}
	canonical, _ := util.CanonicalizePath(filepath.ToSlash(p))
	return canonical, nil
}

// Parse processes cl.exe output, returning the output with /showIncludes
// lines (and the source filename echoed by cl) removed, plus the normalized
// set of non-system include paths.
func Parse(output, depsPrefix, cwd string) (string, []string, error) {
	if depsPrefix == "" {
		depsPrefix = defaultDepsPrefix
	}

	var filtered strings.Builder
	seenShowIncludes := false
	included := map[string]struct{}{}

	start := 0
	for start < len(output) {
		end := strings.IndexAny(output[start:], "\r\n")
		if end < 0 {
			end = len(output)
		} else {
			end += start
		}
		line := output[start:end]

		if include, ok := filterShowIncludes(line, depsPrefix); ok {
			seenShowIncludes = true
			normalized, err := NormalizeInclude(include, cwd)
			if err != nil {
				return "", nil, err
			}
			if !IsSystemInclude(normalized) {
				included[normalized] = struct{}{}
			}
		} else if !seenShowIncludes && isInputFilename(line) {
			// cl.exe echoes the source filename before the includes; ninja
			// suppresses it.
		} else {
			filtered.WriteString(line)
			filtered.WriteByte('\n')
		}

		if end < len(output) && output[end] == '\r' {
			end++
		}
		if end < len(output) && output[end] == '\n' {
			end++
		}
		start = end
	}

	includes := make([]string, 0, len(included))
	for include := range included {
		includes = append(includes, include)
	}
	return filtered.String(), includes, nil
}

func filterShowIncludes(line, prefix string) (string, bool) {
	if !strings.HasPrefix(line, prefix) {
		return "", false
	}
	include := strings.TrimLeft(line[len(prefix):], " ")
	if include == "" {
		return "", false
	}
	return include, true
}

func isInputFilename(line string) bool {
	lower := strings.ToLower(strings.TrimSpace(line))
	for _, ext := range []string{".c", ".cc", ".cxx", ".cpp", ".c++"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}
