//go:build windows

package command_line

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitWindowsQuoting(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    []string
	}{
		{
			name:    "quoted executable path",
			command: `"C:\Program Files\LLVM\bin\clang.exe" -c C:\src\main.c -o C:\out\main.o`,
			want:    []string{`C:\Program Files\LLVM\bin\clang.exe`, "-c", `C:\src\main.c`, "-o", `C:\out\main.o`},
		},
		{
			name:    "quoted argument with spaces",
			command: `cl.exe /I"C:\Program Files\SDK" /c foo.c`,
			want:    []string{"cl.exe", `/IC:\Program Files\SDK`, "/c", "foo.c"},
		},
		{
			name:    "cmd shell command",
			command: `cmd.exe /c "cd /D C:\repo && cl.exe /nologo"`,
			want:    []string{"cmd.exe", "/c", `cd /D C:\repo && cl.exe /nologo`},
		},
		{
			name:    "empty",
			command: "",
			want:    []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Split(tc.command)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
