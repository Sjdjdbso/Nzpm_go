package upload_utils

import (
	"strconv"
	"testing"
)

func TestRcloneProgressRegex(t *testing.T) {
	line := "Transferred:   10.500 MiB / 100.000 MiB, 10%, 2.500 MiB/s, ETA 35s"
	matches := rcloneProgressRegex.FindStringSubmatch(line)
	if len(matches) < 6 {
		t.Fatalf("Expected 6 matches, got %d for line: %s", len(matches), line)
	}

	if matches[1] != "10.500 MiB" {
		t.Errorf("Expected transferred 10.500 MiB, got %s", matches[1])
	}
	if matches[2] != "100.000 MiB" {
		t.Errorf("Expected total 100.000 MiB, got %s", matches[2])
	}
	pct, err := strconv.ParseFloat(matches[3], 64)
	if err != nil || pct != 10 {
		t.Errorf("Expected percentage 10, got %s (err: %v)", matches[3], err)
	}
	if matches[4] != "2.500 MiB/s" {
		t.Errorf("Expected speed 2.500 MiB/s, got %s", matches[4])
	}
	if matches[5] != "35s" {
		t.Errorf("Expected ETA 35s, got %s", matches[5])
	}
}
