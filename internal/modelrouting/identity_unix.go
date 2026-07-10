//go:build !windows

package modelrouting

import (
	"fmt"
	"os"
	"syscall"
)

func fileObjectIdentity(path string) (string, error) {
	information, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	stat, ok := information.Sys().(*syscall.Stat_t)
	if !ok {
		return "", fmt.Errorf("filesystem identity unavailable")
	}
	return fmt.Sprintf("unix:%d:%d", uint64(stat.Dev), uint64(stat.Ino)), nil
}
