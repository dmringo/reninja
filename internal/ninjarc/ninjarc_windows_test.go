//go:build windows

package ninjarc

import (
	"flag"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolvePathWindowsDrive(t *testing.T) {
	workspace := `C:\work\repo`
	assert.Equal(t, "work/repo/.ninjarc", resolvePath(workspace, `%workspace%/.ninjarc`))
	assert.Equal(t, "work/repo/.ninjarc", resolvePath(workspace, `C:\work\repo\.ninjarc`))
	assert.Equal(t, "etc/.ninjarc", resolvePath(workspace, `/etc/.ninjarc`))
}

func TestResolvePathWindowsHome(t *testing.T) {
	current, err := user.Current()
	require.NoError(t, err)
	home := filepath.ToSlash(current.HomeDir)
	home = strings.TrimPrefix(home, filepath.VolumeName(current.HomeDir))
	home = strings.TrimLeft(home, "/")
	assert.Equal(t, home+"/.ninjarc", resolvePath(`C:\work\repo`, `~/.ninjarc`))
}

func TestRCFSRootWindows(t *testing.T) {
	assert.Equal(t, `C:\`, rcFSRoot(`C:\work\repo`))
}

func TestParseRCFilesFromRealFilesystem(t *testing.T) {
	workspace := t.TempDir()
	rcPath := filepath.Join(workspace, ".ninjarc")
	require.NoError(t, os.WriteFile(rcPath, []byte("build --jobs=12\n"), 0644))

	fsys := os.DirFS(rcFSRoot(workspace))
	rc, err := ParseRCFiles(fsys, workspace, `%workspace%/.ninjarc`)
	require.NoError(t, err)

	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	jobs := flags.String("jobs", "", "")
	rc.Apply("build", "", flags, nil)
	assert.Equal(t, "12", *jobs)

	// The explicit --ninjarc path is made absolute by the caller. Absolute
	// paths inside the workspace must resolve through the volume-rooted FS.
	rc, err = ParseRCFiles(fsys, workspace, rcPath)
	require.NoError(t, err)
	flags = flag.NewFlagSet("test", flag.ContinueOnError)
	jobs = flags.String("jobs", "", "")
	rc.Apply("build", "", flags, nil)
	assert.Equal(t, "12", *jobs)
}

func TestParseRCFilesDifferentVolumeIsOptional(t *testing.T) {
	workspace := t.TempDir()
	// A path on a different volume cannot be resolved against the workspace
	// volume root. Optional files are ignored rather than read incorrectly.
	other := `Z:\does-not-exist\.ninjarc`
	rc, err := ParseRCFiles(os.DirFS(rcFSRoot(workspace)), workspace, other)
	require.NoError(t, err)
	require.NotNil(t, rc)
}
