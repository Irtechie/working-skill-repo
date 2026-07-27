//go:build windows

package modelrouting

import (
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var getFinalPathNameByHandleW = syscall.NewLazyDLL("kernel32.dll").NewProc("GetFinalPathNameByHandleW")

const (
	fileNameNormalized = 0
	volumeNameDOS      = 0
)

// canonicalizeExistingPath resolves the actual filesystem object through an
// open handle. filepath.EvalSymlinks can reject valid Windows junction roots,
// even though CreateFile follows the junction and exposes its final target.
// Keeping this handle-based avoids treating a caller-controlled path spelling
// as the project identity source.
func canonicalizeExistingPath(path string) (string, error) {
	pointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}
	handle, err := syscall.CreateFile(
		pointer,
		0,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return "", err
	}
	defer syscall.CloseHandle(handle)

	buffer := make([]uint16, 260)
	for {
		length, _, callErr := syscall.Syscall6(
			getFinalPathNameByHandleW.Addr(),
			4,
			uintptr(handle),
			uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(len(buffer)),
			uintptr(fileNameNormalized|volumeNameDOS),
			0,
			0,
		)
		if length == 0 {
			return "", callErr
		}
		if length < uintptr(len(buffer)) {
			return filepath.Clean(trimExtendedPathPrefix(syscall.UTF16ToString(buffer[:length]))), nil
		}
		buffer = make([]uint16, length+1)
	}
}

func trimExtendedPathPrefix(path string) string {
	const extendedPrefix = `\\?\`
	const extendedUNCPrefix = `\\?\UNC\`
	if strings.HasPrefix(path, extendedUNCPrefix) {
		return `\\` + strings.TrimPrefix(path, extendedUNCPrefix)
	}
	return strings.TrimPrefix(path, extendedPrefix)
}

func fileObjectIdentity(path string) (string, error) {
	pointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}
	handle, err := syscall.CreateFile(
		pointer,
		0,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return "", err
	}
	defer syscall.CloseHandle(handle)
	var information syscall.ByHandleFileInformation
	if err := syscall.GetFileInformationByHandle(handle, &information); err != nil {
		return "", err
	}
	return fmt.Sprintf("windows:%08x:%08x%08x", information.VolumeSerialNumber, information.FileIndexHigh, information.FileIndexLow), nil
}
