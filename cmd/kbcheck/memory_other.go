//go:build !windows

package main

import (
	"os"
	"strconv"
	"strings"
)

// availableProcessMemoryBytes reports memory available to new processes.
// Linux exposes MemAvailable directly; other platforms report unknown so the
// caller falls back to its conservative default rather than guessing.
func availableProcessMemoryBytes() (uint64, bool) {
	content, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, false
	}
	for _, line := range strings.Split(string(content), "\n") {
		key, value, found := strings.Cut(line, ":")
		if !found || key != "MemAvailable" {
			continue
		}
		fields := strings.Fields(value)
		if len(fields) == 0 {
			return 0, false
		}
		kb, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil || kb == 0 {
			return 0, false
		}
		return kb * 1024, true
	}
	return 0, false
}
