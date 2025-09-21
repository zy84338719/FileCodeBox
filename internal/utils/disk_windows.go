//go:build windows

package utils

import (
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// GetUsagePercent returns disk usage percentage for the drive containing the given path on Windows.
func GetUsagePercent(path string) (float64, error) {
	volume, err := resolveVolume(path)
	if err != nil {
		return 0, err
	}

	var (
		freeBytesAvailable uint64
		totalNumberOfBytes uint64
		totalNumberOfFree  uint64
	)

	if err := windows.GetDiskFreeSpaceEx(windows.StringToUTF16Ptr(volume), &freeBytesAvailable, &totalNumberOfBytes, &totalNumberOfFree); err != nil {
		return 0, err
	}

	if totalNumberOfBytes == 0 {
		return 0, fmt.Errorf("unable to compute total disk size for %s", volume)
	}

	used := totalNumberOfBytes - totalNumberOfFree
	usage := (float64(used) / float64(totalNumberOfBytes)) * 100.0
	return usage, nil
}

func resolveVolume(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	volume := filepath.VolumeName(absPath)
	if volume == "" {
		return "", fmt.Errorf("unable to determine volume for path %s", absPath)
	}

	// Ensure the volume points to root (e.g., "C:\")
	if !strings.HasSuffix(volume, "\\") {
		volume += "\\"
	}

	return volume, nil
}
