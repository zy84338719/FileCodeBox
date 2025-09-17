package utils

import (
	"os"
	"runtime"
	"testing"
)

// Tests rely on syscall.Statfs which is not supported on Windows in this codepath.
func TestGetUsagePercent_ValidPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping disk usage test on windows")
	}

	dir, err := os.MkdirTemp("", "disktest")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	usage, err := GetUsagePercent(dir)
	if err != nil {
		t.Fatalf("expected no error for valid path, got: %v", err)
	}
	if usage < 0 || usage > 100 {
		t.Fatalf("usage percent out of range: %v", usage)
	}
}

func TestGetUsagePercent_InvalidPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping disk usage test on windows")
	}

	// choose a path that almost certainly doesn't exist
	path := "/path/that/does/not/exist_for_disk_test"
	_, err := GetUsagePercent(path)
	if err == nil {
		t.Fatalf("expected error for invalid path, got nil")
	}
}
