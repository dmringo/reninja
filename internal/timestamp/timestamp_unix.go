//go:build !windows

package timestamp

import "time"

// FromTime converts a file modification time to the timestamp representation
// used in the build and deps logs. On POSIX this matches ninja: nanoseconds
// since the Unix epoch.
func FromTime(t time.Time) TimeStamp {
	return TimeStamp(t.UnixNano())
}
