package clparser_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/buildbuddy-io/reninja/internal/clparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFiltersShowIncludes(t *testing.T) {
	cwd := t.TempDir()
	localInclude := filepath.Join(cwd, "..", "src", "a.h")
	systemInclude := filepath.Join(cwd, "Program Files", "SDK", "b.h")

	output := strings.Join([]string{
		"main.c",
		"Note: including file:  " + localInclude,
		"Note: including file: " + systemInclude,
		"warning: something happened",
		"",
	}, "\r\n")

	filtered, includes, err := clparser.Parse(output, "Note: including file: ", cwd)
	require.NoError(t, err)

	assert.NotContains(t, filtered, "main.c")
	assert.NotContains(t, filtered, "including file")
	assert.Contains(t, filtered, "warning: something happened")
	assert.ElementsMatch(t, []string{"../src/a.h"}, includes)
}

func TestParseDefaultPrefix(t *testing.T) {
	cwd := t.TempDir()
	include := filepath.Join(cwd, "a.h")
	_, includes, err := clparser.Parse("main.c\nNote: including file: "+include+"\n", "", cwd)
	require.NoError(t, err)
	assert.Equal(t, []string{"a.h"}, includes)
}

func TestParsePreservesSourceLookingLinesAfterIncludes(t *testing.T) {
	cwd := t.TempDir()
	include := filepath.Join(cwd, "a.h")
	filtered, _, err := clparser.Parse(
		"main.c\nNote: including file: "+include+"\nfoo.c\n",
		"Note: including file: ",
		cwd,
	)
	require.NoError(t, err)
	assert.Contains(t, filtered, "foo.c")
}

func TestParseDeduplicatesIncludes(t *testing.T) {
	cwd := t.TempDir()
	include := filepath.Join(cwd, "a.h")
	_, includes, err := clparser.Parse(
		"Note: including file: "+include+"\nNote: including file: "+include+"\n",
		"Note: including file: ",
		cwd,
	)
	require.NoError(t, err)
	assert.Equal(t, []string{"a.h"}, includes)
}

func TestParseEmptyIncludeIsKeptAsOutput(t *testing.T) {
	cwd := t.TempDir()
	filtered, includes, err := clparser.Parse("Note: including file: \n", "Note: including file: ", cwd)
	require.NoError(t, err)
	assert.Empty(t, includes)
	assert.Contains(t, filtered, "Note: including file:")
}

func TestNormalizeInclude(t *testing.T) {
	cwd := filepath.Join(t.TempDir(), "build")
	include := filepath.Join(cwd, "..", "src", "a.h")
	got, err := clparser.NormalizeInclude(include, cwd)
	require.NoError(t, err)
	assert.Equal(t, "../src/a.h", got)

	got, err = clparser.NormalizeInclude("relative.h", cwd)
	require.NoError(t, err)
	assert.Equal(t, "relative.h", got)
}

func TestIsSystemInclude(t *testing.T) {
	assert.True(t, clparser.IsSystemInclude(`C:\Program Files\SDK\a.h`))
	assert.True(t, clparser.IsSystemInclude(`C:\Microsoft Visual Studio\VC\a.h`))
	assert.False(t, clparser.IsSystemInclude(`C:\repo\src\a.h`))
}
