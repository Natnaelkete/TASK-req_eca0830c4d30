//go:build !windows

package services

import (
	"fmt"
	"os"
	"syscall"
)

// getDiskUsagePercent returns the disk usage percentage for the root filesystem.
func getDiskUsagePercent() (float64, error) {
	path := "/"
	if _, err := os.Stat("/app"); err == nil {
		path = "/app"
	}

	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, fmt.Errorf("statfs %s: %w", path, err)
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	if total == 0 {
		return 0, nil
	}

	used := total - free
	return float64(used) / float64(total) * 100, nil
}
