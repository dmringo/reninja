//go:build windows

package util

import (
	"os"

	"github.com/shirou/gopsutil/v4/load"
)

func GetLoadAverage() float64 {
	avg, err := load.Avg()
	if err != nil {
		return -1
	}
	return avg.Load1
}

// ReplaceFileContent replaces the content of destPath with the content of
// tempPath by removing destPath and renaming tempPath to destPath.
// On Windows, file ownership is handled differently and is not preserved.
func ReplaceFileContent(destPath, tempPath string) error {
	if err := os.Remove(destPath); err != nil {
		return err
	}

	if err := os.Rename(tempPath, destPath); err != nil {
		return err
	}

	return nil
}
