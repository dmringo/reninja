//go:build !windows

package disk

import "os"

func removeFile(path string) error {
	return os.Remove(path)
}
