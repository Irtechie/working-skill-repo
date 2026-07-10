//go:build windows

package modelrouting

import (
	"fmt"
	"syscall"
)

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
