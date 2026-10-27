//go:build windows

package subprocess

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

// newCommandContext mirrors ninja's Windows behavior: the command string is
// passed to CreateProcess as-is (no "cmd /c" wrapper), and the executable is
// resolved from its first token.
func newCommandContext(ctx context.Context, command string, useConsole bool) (*exec.Cmd, error) {
	argv, err := windows.DecomposeCommandLine(command)
	if err != nil {
		return nil, fmt.Errorf("parsing command line: %w", err)
	}
	if len(argv) == 0 || argv[0] == "" {
		return nil, fmt.Errorf("empty command line")
	}

	cmd := exec.CommandContext(ctx, argv[0])
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: command,
	}
	// Match ninja: non-console subprocesses get their own process group so
	// they can be interrupted independently.
	if !useConsole {
		cmd.SysProcAttr.CreationFlags = windows.CREATE_NEW_PROCESS_GROUP
	}
	return cmd, nil
}

func interruptSignals() []os.Signal {
	// Go maps both Ctrl-C and Ctrl-Break to os.Interrupt on Windows.
	return []os.Signal{os.Interrupt, syscall.SIGTERM}
}

// CONTROL_C_EXIT is the exit code Windows uses for processes terminated by
// Ctrl-C.
const controlCExit = 0xC000013A

func isInterruptExitCode(code int) bool {
	return uint32(code) == controlCExit
}
