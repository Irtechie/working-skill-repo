//go:build !windows

package modelrouting

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestSharedProjectLockCreationHonorsProjectUmask(t *testing.T) {
	previousUmask := syscall.Umask(0)
	defer syscall.Umask(previousUmask)

	root := filepath.Join(t.TempDir(), "shared")
	lock, err := AcquireSharedProjectLock(root, "work-queue.lock", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(root, "work-queue.lock"))
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o666 {
		t.Fatalf("shared lock mode=%#o want %#o", got, os.FileMode(0o666))
	}
}
