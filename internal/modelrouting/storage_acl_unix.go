//go:build !windows

package modelrouting

import (
	"os"
	"syscall"
)

func secureStorageDirectorySecurity(path string) error {
	if err := validateStorageOwnerAndType(path, true, false); err != nil {
		return err
	}
	if err := os.Chmod(path, 0o700); err != nil {
		return err
	}
	return validateStorageDirectorySecurity(path)
}

func validateStorageDirectorySecurity(path string) error {
	return validateStorageOwnerAndType(path, true, true)
}

func secureStorageFileSecurity(path string) error {
	if err := validateStorageOwnerAndType(path, false, false); err != nil {
		return err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return err
	}
	return validateStorageFileSecurity(path)
}

func validateStorageFileSecurity(path string) error {
	return validateStorageOwnerAndType(path, false, true)
}

func validateStorageOwnerAndType(path string, directory, strictMode bool) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || directory != info.IsDir() || (!directory && !info.Mode().IsRegular()) {
		return ErrUnsafePath
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || int(stat.Uid) != os.Geteuid() {
		return ErrUnsafePath
	}
	if strictMode && info.Mode().Perm()&0o077 != 0 {
		return ErrUnsafePath
	}
	return nil
}
