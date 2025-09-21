package utils

import (
	"fmt"
	"syscall"
)

// GetUsagePercent attempts to get disk usage percent for a given path (0-100).
// Returns error on unsupported platforms or when statfs fails.
func GetUsagePercent(path string) (float64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}

	total := float64(stat.Blocks) * float64(stat.Bsize)
	free := float64(stat.Bfree) * float64(stat.Bsize)
	used := total - free
	if total <= 0 {
		return 0, fmt.Errorf("unable to compute total disk size")
	}

	usage := (used / total) * 100.0
	return usage, nil
}
