// Package command_line splits command lines into arguments using the
// platform's native quoting rules.
package command_line

// Split splits a command line into arguments. On Windows it follows
// CreateProcess/CommandLineToArgvW rules; elsewhere it uses POSIX shell rules.
func Split(command string) ([]string, error) {
	return split(command)
}
