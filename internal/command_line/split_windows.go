//go:build windows

package command_line

import "golang.org/x/sys/windows"

func split(command string) ([]string, error) {
	return windows.DecomposeCommandLine(command)
}
