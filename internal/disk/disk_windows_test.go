//go:build windows

package disk_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/buildbuddy-io/reninja/internal/disk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteFileCRLF(t *testing.T) {
	dir := t.TempDir()
	d := disk.NewRealDiskInterface()

	crlfPath := filepath.Join(dir, "crlf.rsp")
	require.NoError(t, d.WriteFile(crlfPath, []byte("a\nb\n"), true /*=crlfOnWindows*/))
	contents, err := os.ReadFile(crlfPath)
	require.NoError(t, err)
	assert.Equal(t, "a\r\nb\r\n", string(contents))

	lfPath := filepath.Join(dir, "lf.rsp")
	require.NoError(t, d.WriteFile(lfPath, []byte("a\nb\n"), false /*=crlfOnWindows*/))
	contents, err = os.ReadFile(lfPath)
	require.NoError(t, err)
	assert.Equal(t, "a\nb\n", string(contents))
}

func TestRemoveReadOnlyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "readonly.txt")
	require.NoError(t, os.WriteFile(path, []byte("x"), 0644))
	require.NoError(t, os.Chmod(path, 0444))

	d := disk.NewRealDiskInterface()
	assert.Equal(t, 0, d.RemoveFile(path))
	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

func TestRemoveMissingFile(t *testing.T) {
	d := disk.NewRealDiskInterface()
	assert.Equal(t, 1, d.RemoveFile(filepath.Join(t.TempDir(), "missing.txt")))
}
