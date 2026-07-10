//go:build windows

package modelrouting

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	lockfileFailImmediately = 0x00000001
	lockfileExclusiveLock   = 0x00000002
	errorLockViolation      = syscall.Errno(33)
)

var (
	lockFileEx   = storageKernel32.NewProc("LockFileEx")
	unlockFileEx = storageKernel32.NewProc("UnlockFileEx")
)

func lockStorageFile(file *os.File) error {
	overlapped := new(syscall.Overlapped)
	result, _, callErr := lockFileEx.Call(file.Fd(), lockfileExclusiveLock, 0, 1, 0, uintptr(unsafe.Pointer(overlapped)))
	if result == 0 {
		return fmt.Errorf("lock private state: %w", callErr)
	}
	return nil
}

func tryLockStorageFile(file *os.File) (bool, error) {
	overlapped := new(syscall.Overlapped)
	result, _, callErr := lockFileEx.Call(file.Fd(), lockfileExclusiveLock|lockfileFailImmediately, 0, 1, 0, uintptr(unsafe.Pointer(overlapped)))
	if result != 0 {
		return true, nil
	}
	if errno, ok := callErr.(syscall.Errno); ok && errno == errorLockViolation {
		return false, nil
	}
	return false, fmt.Errorf("try lock private state: %w", callErr)
}

func unlockStorageFile(file *os.File) {
	overlapped := new(syscall.Overlapped)
	_, _, _ = unlockFileEx.Call(file.Fd(), 0, 1, 0, uintptr(unsafe.Pointer(overlapped)))
}
