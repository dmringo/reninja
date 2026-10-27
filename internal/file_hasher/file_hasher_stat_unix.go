//go:build !windows

package file_hasher

import (
	"os"
	"syscall"

	"github.com/djherbis/times"
)

func statKey(path string, fi os.FileInfo, t times.Timespec, digestType int32) (hashKey, bool) {
	stat, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return hashKey{}, false
	}
	return hashKey{
		dev:        uint64(stat.Dev),
		inode:      uint64(stat.Ino),
		ctimeNano:  t.ChangeTime().UnixNano(),
		digestType: digestType,
	}, true
}
