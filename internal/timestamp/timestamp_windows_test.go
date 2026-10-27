//go:build windows

package timestamp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFromTimeMatchesNinjaWindowsEpoch(t *testing.T) {
	// Ninja's Windows epoch is not exactly 2001-01-01: it subtracts
	// 12622770400 seconds from FILETIME, which places zero at
	// 2001-01-01 00:00:00 UTC minus 10400 seconds.
	epoch := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, TimeStamp(104_000_000_000), FromTime(epoch))
	assert.Equal(t, TimeStamp(104_010_000_000), FromTime(epoch.Add(time.Second)))
	assert.Equal(t, TimeStamp(104_012_345_678), FromTime(epoch.Add(1*time.Second+234_567_800*time.Nanosecond)))
}

func TestFromTimeBeforeNinjaWindowsEpoch(t *testing.T) {
	assert.Negative(t, FromTime(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)))
}
