//go:build !windows

package subprocess

import (
	"context"
	"os"
	"os/exec"
	"syscall"
)

func interruptSignals() []os.Signal {
	return []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGHUP}
}

func newCommandContext(ctx context.Context, command string, useConsole bool) (*exec.Cmd, error) {
	return exec.CommandContext(ctx, "/bin/sh", "-c", command), nil
}

func isInterruptExitCode(code int) bool {
	return false
}
