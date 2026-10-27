package subprocess_test

import (
	"testing"

	"github.com/buildbuddy-io/reninja/internal/exit_status"
	"github.com/buildbuddy-io/reninja/internal/subprocess"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// waitFor runs the subprocess set until the given subprocess completes.
func waitFor(subprocs *subprocess.Set, subproc *subprocess.Subprocess) {
	for !subproc.Done() {
		subprocs.DoWork()
	}
	subprocs.DoWork()
}

func TestBadCommandStderr(t *testing.T) {
	subprocs := subprocess.NewSet()
	subproc, err := subprocs.Add("cmd /c ninja_no_such_command", false /*=useConsole*/)
	require.NoError(t, err)
	require.NotNil(t, subproc)

	waitFor(subprocs, subproc)

	assert.NotEqual(t, exit_status.ExitSuccess, subproc.Finish())
}

func TestNoSuchCommand(t *testing.T) {
	subprocs := subprocess.NewSet()
	subproc, err := subprocs.Add("ninja_no_such_command", false /*=useConsole*/)
	require.NoError(t, err)
	require.NotNil(t, subproc)

	waitFor(subprocs, subproc)

	assert.NotEqual(t, exit_status.ExitSuccess, subproc.Finish())
}
