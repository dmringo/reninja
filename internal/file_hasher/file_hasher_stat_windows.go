//go:build windows

package file_hasher

import (
	"os"

	"github.com/djherbis/times"
)

func statKey(path string, fi os.FileInfo, t times.Timespec, digestType int32) (hashKey, bool) {
	// os.FileInfo on Windows does not expose a stable file ID, so use the
	// path together with the size and mtime.
	return hashKey{
		path:       path,
		size:       fi.Size(),
		ctimeNano:  fi.ModTime().UnixNano(),
		digestType: digestType,
	}, true
}
