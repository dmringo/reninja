//go:build windows

package file_hasher_test

import (
	"os"
	"testing"
	"time"

	"github.com/buildbuddy-io/reninja/internal/file_hasher"
	"github.com/stretchr/testify/require"

	repb "github.com/buildbuddy-io/reninja/genproto/remote_execution"
)

func TestHashFileCacheInvalidatedByMtimeChange(t *testing.T) {
	path := writeFile(t, "cache_windows.txt", "aaaa")

	first, err := file_hasher.HashFile(path, repb.DigestFunction_SHA256)
	require.NoError(t, err)

	// Same size, different contents. Explicitly advance the mtime so the
	// Windows cache key changes even if the filesystem timestamp granularity
	// would otherwise hide the rewrite.
	require.NoError(t, os.WriteFile(path, []byte("bbbb"), 0644))
	future := time.Now().Add(2 * time.Second)
	require.NoError(t, os.Chtimes(path, future, future))

	second, err := file_hasher.HashFile(path, repb.DigestFunction_SHA256)
	require.NoError(t, err)
	require.NotEqual(t, first, second)
}
