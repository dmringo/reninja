//go:build !windows

package line_printer

import (
	"os"

	"github.com/mattn/go-isatty"
)

func isSmartTerminal() bool {
	// A smart terminal supports escape sequences. Check that stdout is a tty
	// and that TERM is set (even if empty) and not "dumb".
	// This matches C++ ninja behavior: TERM="" is smart, TERM not set is not.
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		return false
	}
	term, isSet := os.LookupEnv("TERM")
	if !isSet {
		return false
	}
	return term != "dumb"
}
