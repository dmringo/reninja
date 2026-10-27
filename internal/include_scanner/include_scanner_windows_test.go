//go:build windows

package include_scanner_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildbuddy-io/reninja/internal/include_scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanEdgeQuotedSearchPathWithSpaces(t *testing.T) {
	dir := t.TempDir()
	incDir := filepath.Join(dir, "include dir")
	require.NoError(t, os.MkdirAll(incDir, 0755))

	header := filepath.Join(incDir, "sys.h")
	require.NoError(t, os.WriteFile(header, []byte(""), 0644))

	src := filepath.Join(dir, "main.c")
	require.NoError(t, os.WriteFile(src, []byte("#include <sys.h>\n"), 0644))

	command := fmt.Sprintf(`clang -I"%s" "%s"`, incDir, src)
	extra, err := include_scanner.New().ScanEdge([]string{src}, command)
	require.NoError(t, err)
	assert.Contains(t, absPaths(t, extra), header)
}

func TestExtractIntermediateDirsFromWindowsPath(t *testing.T) {
	dirs := include_scanner.ExtractIntermediateDirsFromCommand(`clang -IC:\repo\build\..\include foo.c`)
	assert.Equal(t, []string{"C:/repo/build"}, dirs)
}

func TestExtractCommandReferencedPathsSeparatorMismatch(t *testing.T) {
	root := t.TempDir()
	script := filepath.Join(root, "scripts", "build.py")
	require.NoError(t, os.MkdirAll(filepath.Dir(script), 0755))
	require.NoError(t, os.WriteFile(script, []byte("print(1)\n"), 0644))

	// Build files commonly use forward slashes even on Windows.
	command := "python " + filepath.ToSlash(script)
	paths, err := include_scanner.ExtractCommandReferencedPaths(command, root)
	require.NoError(t, err)
	assert.Contains(t, paths, script)
}
