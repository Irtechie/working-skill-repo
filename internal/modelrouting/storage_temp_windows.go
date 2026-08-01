//go:build windows

package modelrouting

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

var windowsCreateFile = syscall.CreateFile

func createStorageTempFile(parent string, private bool) (*os.File, error) {
	if !private {
		return os.CreateTemp(parent, ".catalog-*.tmp")
	}
	sid, err := currentWindowsSID()
	if err != nil {
		return nil, err
	}
	descriptor, free, err := windowsDescriptor(expectedWindowsStorageSDDL(sid, false))
	if err != nil {
		return nil, err
	}
	defer free()
	security := syscall.SecurityAttributes{
		Length:             uint32(unsafe.Sizeof(syscall.SecurityAttributes{})),
		SecurityDescriptor: descriptor,
	}
	for attempt := 0; attempt < 100; attempt++ {
		suffix := make([]byte, 16)
		if _, err := rand.Read(suffix); err != nil {
			return nil, err
		}
		name := filepath.Join(parent, ".catalog-"+hex.EncodeToString(suffix)+".tmp")
		namePointer, err := syscall.UTF16PtrFromString(name)
		if err != nil {
			return nil, err
		}
		handle, err := windowsCreateFile(
			namePointer,
			syscall.GENERIC_READ|syscall.GENERIC_WRITE,
			0,
			&security,
			syscall.CREATE_NEW,
			syscall.FILE_ATTRIBUTE_NORMAL,
			0,
		)
		if err == nil {
			return os.NewFile(uintptr(handle), name), nil
		}
		if !errors.Is(err, syscall.ERROR_FILE_EXISTS) && !errors.Is(err, syscall.ERROR_ALREADY_EXISTS) {
			return nil, err
		}
	}
	return nil, os.ErrExist
}
