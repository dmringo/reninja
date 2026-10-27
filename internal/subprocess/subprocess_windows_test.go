//go:build windows

package subprocess

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buildbuddy-io/reninja/internal/exit_status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func addWindowsCommand(t *testing.T, set *Set, command string) *Subprocess {
	t.Helper()
	subproc, err := set.Add(command, false /*=useConsole*/)
	require.NoError(t, err)
	require.NotNil(t, subproc)
	for !subproc.Done() {
		set.DoWork()
	}
	set.DoWork()
	return subproc
}

func TestWindowsCommandLineIsNotRewritten(t *testing.T) {
	cmd, err := newCommandContext(context.Background(), `foo.exe "bar baz" C:\qux`, false /*=useConsole*/)
	require.NoError(t, err)
	require.NotNil(t, cmd.SysProcAttr)
	assert.Equal(t, `foo.exe "bar baz" C:\qux`, cmd.SysProcAttr.CmdLine)
	assert.NotZero(t, cmd.SysProcAttr.CreationFlags&windows.CREATE_NEW_PROCESS_GROUP)

	cmd, err = newCommandContext(context.Background(), `foo.exe`, true /*=useConsole*/)
	require.NoError(t, err)
	assert.Zero(t, cmd.SysProcAttr.CreationFlags&windows.CREATE_NEW_PROCESS_GROUP)
}

func TestWindowsShellBuiltinsRequireCmd(t *testing.T) {
	set := NewSet()
	subproc, err := set.Add(`cd /d C:\`, false /*=useConsole*/)
	require.NoError(t, err)
	for !subproc.Done() {
		set.DoWork()
	}
	set.DoWork()
	assert.NotEqual(t, exit_status.ExitSuccess, subproc.Finish(),
		"CreateProcess should not fall back to cmd.exe for shell builtins")
}

func TestWindowsExplicitCmdShell(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "out.txt")
	set := NewSet()
	subproc := addWindowsCommand(t, set, fmt.Sprintf(`cmd /c echo hello > "%s"`, outPath))
	assert.Equal(t, exit_status.ExitSuccess, subproc.Finish())

	contents, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Contains(t, string(contents), "hello")
}

func TestWindowsQuotedExecutableWithSpaces(t *testing.T) {
	comspec := os.Getenv("COMSPEC")
	if comspec == "" {
		comspec = `C:\Windows\System32\cmd.exe`
	}
	if _, err := os.Stat(comspec); err != nil {
		t.Skipf("cmd.exe not available: %s", err)
	}

	dir := filepath.Join(t.TempDir(), "path with spaces")
	require.NoError(t, os.MkdirAll(dir, 0755))
	quotedExe := filepath.Join(dir, "cmd copy.exe")
	exeBytes, err := os.ReadFile(comspec)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(quotedExe, exeBytes, 0755))

	set := NewSet()
	subproc := addWindowsCommand(t, set, fmt.Sprintf(`"%s" /c echo quoted-ok`, quotedExe))
	assert.Equal(t, exit_status.ExitSuccess, subproc.Finish())
	assert.Contains(t, subproc.GetOutput(), "quoted-ok")
}

func TestWindowsPreservesBackslashesAndQuotedArguments(t *testing.T) {
	set := NewSet()
	subproc := addWindowsCommand(t, set, `cmd /c echo C:\foo\bar "a b"`)
	assert.Equal(t, exit_status.ExitSuccess, subproc.Finish())
	output := subproc.GetOutput()
	assert.Contains(t, output, `C:\foo\bar`)
	assert.Contains(t, output, "a b")
}

func TestWindowsCapturesStdoutAndStderr(t *testing.T) {
	set := NewSet()
	subproc := addWindowsCommand(t, set, `cmd /c echo out & echo err 1>&2`)
	assert.Equal(t, exit_status.ExitSuccess, subproc.Finish())
	output := subproc.GetOutput()
	assert.Contains(t, output, "out")
	assert.Contains(t, output, "err")
}

func TestWindowsExitCode(t *testing.T) {
	set := NewSet()
	subproc := addWindowsCommand(t, set, `cmd /c exit 7`)
	assert.Equal(t, exit_status.ExitStatusType(7), subproc.Finish())
}

func TestWindowsControlCExitCodeMapping(t *testing.T) {
	assert.True(t, isInterruptExitCode(int(uint32(0xC000013A))))
	assert.False(t, isInterruptExitCode(0))
	assert.False(t, isInterruptExitCode(1))
}

func TestWindowsKillTerminatesChild(t *testing.T) {
	set := NewSet()
	subproc, err := set.Add(`cmd /c ping -n 10 127.0.0.1`, false /*=useConsole*/)
	require.NoError(t, err)
	require.NotNil(t, subproc)

	subproc.Kill()
	require.True(t, subproc.TryFinish(true /*=wait*/))
	assert.NotEqual(t, exit_status.ExitSuccess, subproc.Finish())
}

func TestWindowsOutputIsNotInterpolated(t *testing.T) {
	set := NewSet()
	subproc := addWindowsCommand(t, set, `cmd /c echo %PATH%`)
	assert.Equal(t, exit_status.ExitSuccess, subproc.Finish())
	assert.False(t, strings.Contains(subproc.GetOutput(), "%PATH%"), "cmd should expand environment variables")
}
