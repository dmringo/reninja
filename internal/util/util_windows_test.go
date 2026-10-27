//go:build windows

package util_test

import (
	"testing"

	"github.com/buildbuddy-io/reninja/internal/graph"
	"github.com/buildbuddy-io/reninja/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestCanonicalizeWindowsBackslashes(t *testing.T) {
	canonical, slashBits := util.CanonicalizePath(`C:\foo\bar`)
	assert.Equal(t, "C:/foo/bar", canonical)
	assert.Equal(t, uint64(0b11), slashBits)

	node := graph.NewNode(canonical, slashBits)
	assert.Equal(t, `C:\foo\bar`, node.PathDecanonicalized())
}

func TestCanonicalizeWindowsMixedSeparators(t *testing.T) {
	canonical, slashBits := util.CanonicalizePath(`C:\foo/bar\baz`)
	assert.Equal(t, "C:/foo/bar/baz", canonical)

	node := graph.NewNode(canonical, slashBits)
	assert.Equal(t, `C:\foo/bar\baz`, node.PathDecanonicalized())
}

func TestCanonicalizeWindowsForwardSlashes(t *testing.T) {
	canonical, slashBits := util.CanonicalizePath(`C:/foo/../bar`)
	assert.Equal(t, "C:/bar", canonical)
	assert.Zero(t, slashBits)

	node := graph.NewNode(canonical, slashBits)
	assert.Equal(t, "C:/bar", node.PathDecanonicalized())
}

func TestCanonicalizeWindowsUNCPath(t *testing.T) {
	canonical, _ := util.CanonicalizePath(`\\server\share\dir\..\file`)
	assert.Equal(t, "//server/share/file", canonical)
}

func TestCanonicalizeWindowsDotSegments(t *testing.T) {
	canonical, _ := util.CanonicalizePath(`C:\a\.\b\..\c`)
	assert.Equal(t, "C:/a/c", canonical)
}

func TestWin32Escaping(t *testing.T) {
	assert.Equal(t, `C:\foo\bar`, util.GetWin32EscapedString(`C:\foo\bar`))
	assert.Equal(t, `"C:\Program Files\foo.exe"`, util.GetWin32EscapedString(`C:\Program Files\foo.exe`))
	assert.Equal(t, `"C:\path with space\\"`, util.GetWin32EscapedString(`C:\path with space\`))
}
