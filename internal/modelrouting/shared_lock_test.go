package modelrouting

import (
	"os"
	"testing"
	"time"
)

func TestSharedProjectLockProductionProbe(t *testing.T) {
	root := os.Getenv("KB_SHARED_LOCK_PROBE_DIR")
	if root == "" {
		t.Skip("KB_SHARED_LOCK_PROBE_DIR is unset")
	}
	lock, err := AcquireSharedProjectLock(root, "work-queue.lock", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
}
