//go:build windows

package timestamp

import "time"

// Ninja's Windows implementation stores file times as 100-nanosecond ticks
// relative to the year 2000 (FILETIME minus the 1600->2000 epoch delta), not
// as Unix nanoseconds. Match that so .ninja_log and .ninja_deps are shared
// with ninja.
const (
	filetimeTicksPerNanosecond = 100
	// Seconds between 1970-01-01 (the Unix epoch) and 2000-01-01 (ninja's
	// Windows epoch): 12622770400 - 11644473600.
	unixEpochToNinjaWindowsEpochSeconds = 978296800
)

func FromTime(t time.Time) TimeStamp {
	return TimeStamp(t.UnixNano()/filetimeTicksPerNanosecond -
		unixEpochToNinjaWindowsEpochSeconds*(1000000000/filetimeTicksPerNanosecond))
}
