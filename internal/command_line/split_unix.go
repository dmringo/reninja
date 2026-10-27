//go:build !windows

package command_line

import "github.com/google/shlex"

func split(command string) ([]string, error) {
	return shlex.Split(command)
}
