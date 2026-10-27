package line_printer

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmartTerminalRedirectedOutput(t *testing.T) {
	readEnd, writeEnd, err := os.Pipe()
	require.NoError(t, err)
	defer readEnd.Close()
	defer writeEnd.Close()

	original := os.Stdout
	os.Stdout = writeEnd
	defer func() { os.Stdout = original }()

	assert.False(t, isSmartTerminal(), "a pipe is not a console")
}

func TestSupportsColorForce(t *testing.T) {
	t.Setenv("CLICOLOR_FORCE", "1")
	assert.True(t, SupportsColor())
}
