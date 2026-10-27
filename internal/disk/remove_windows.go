//go:build windows

package disk

import (
	"os"

	"golang.org/x/sys/windows"
)

func removeFile(path string) error {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	if attrs, err := windows.GetFileAttributes(p); err == nil && attrs&windows.FILE_ATTRIBUTE_READONLY != 0 {
		// Non-Windows remove() happily deletes read-only files. Mirror that
		// behavior so clean and interrupted-build cleanup work consistently.
		_ = windows.SetFileAttributes(p, attrs&^windows.FILE_ATTRIBUTE_READONLY)
	}
	return os.Remove(path)
}
