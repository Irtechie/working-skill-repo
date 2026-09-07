//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

func TestMain(m *testing.M) {
	root := filepath.Join(os.TempDir(), "kbcheck-private-acl-probe")
	lock, err := modelrouting.AcquirePrivateStateLock(root, "probe.lock", 100*time.Millisecond)
	if err == nil {
		_ = lock.Close()
		code := m.Run()
		if err := cleanupPortableRecoveryInstall(); err != nil {
			fmt.Fprintf(os.Stderr, "packed recovery fixture cleanup: %v\n", err)
			code = 1
		}
		os.Exit(code)
	}
	if errors.Is(err, modelrouting.ErrUnsafePath) || strings.Contains(err.Error(), "Access is denied") {
		fmt.Fprintf(os.Stderr, "skipping cmd/kbcheck Windows tests: private ACL setup unavailable: %v\n", err)
		os.Exit(0)
	}
	fmt.Fprintf(os.Stderr, "private ACL probe failed unexpectedly: %v\n", err)
	os.Exit(1)
}
