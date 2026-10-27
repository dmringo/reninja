//go:build windows

package line_printer

import (
	"os"

	"golang.org/x/sys/windows"
)

func isSmartTerminal() bool {
	// Ninja treats any Windows console as a smart terminal; it does not
	// consult TERM.
	var mode uint32
	return windows.GetConsoleMode(windows.Handle(os.Stdout.Fd()), &mode) == nil
}
