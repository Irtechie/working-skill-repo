//go:build windows

package modelrouting

import (
	"fmt"
	"os"
	"os/user"
	"strings"
	"syscall"
	"unsafe"
)

const (
	securityDescriptorRevision        = 1
	ownerSecurityInformation          = 0x00000001
	daclSecurityInformation           = 0x00000004
	protectedDACLSSecurityInformation = 0x80000000
	errorInsufficientBuffer           = syscall.Errno(122)
)

var (
	storageAdvapi32                     = syscall.NewLazyDLL("advapi32.dll")
	storageKernel32                     = syscall.NewLazyDLL("kernel32.dll")
	convertStringSecurityDescriptorToSD = storageAdvapi32.NewProc("ConvertStringSecurityDescriptorToSecurityDescriptorW")
	convertSecurityDescriptorToStringSD = storageAdvapi32.NewProc("ConvertSecurityDescriptorToStringSecurityDescriptorW")
	getFileSecurityW                    = storageAdvapi32.NewProc("GetFileSecurityW")
	setFileSecurityW                    = storageAdvapi32.NewProc("SetFileSecurityW")
	storageLocalFree                    = storageKernel32.NewProc("LocalFree")
)

func secureStorageDirectorySecurity(path string) error {
	return secureWindowsStoragePath(path, true)
}

func validateStorageDirectorySecurity(path string) error {
	return validateWindowsStoragePath(path, true)
}

func secureStorageFileSecurity(path string) error {
	return secureWindowsStoragePath(path, false)
}

func validateStorageFileSecurity(path string) error {
	return validateWindowsStoragePath(path, false)
}

func secureWindowsStoragePath(path string, directory bool) error {
	if err := validateWindowsStorageType(path, directory); err != nil {
		return err
	}
	owner, err := windowsStorageOwner(path)
	if err != nil {
		return err
	}
	sid, err := currentWindowsSID()
	if err != nil || !strings.EqualFold(owner, "O:"+sid) {
		return ErrUnsafePath
	}
	descriptor, free, err := windowsDescriptor(expectedWindowsStorageSDDL(sid, directory))
	if err != nil {
		return err
	}
	defer free()
	pathPointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	information := uintptr(ownerSecurityInformation | daclSecurityInformation | protectedDACLSSecurityInformation)
	result, _, callErr := setFileSecurityW.Call(uintptr(unsafe.Pointer(pathPointer)), information, descriptor)
	if result == 0 {
		return fmt.Errorf("set private Windows ACL: %w", callErr)
	}
	return validateWindowsStoragePath(path, directory)
}

func validateWindowsStoragePath(path string, directory bool) error {
	if err := validateWindowsStorageType(path, directory); err != nil {
		return err
	}
	sid, err := currentWindowsSID()
	if err != nil {
		return err
	}
	actualDescriptor, err := getWindowsFileDescriptor(path, ownerSecurityInformation|daclSecurityInformation)
	if err != nil {
		return err
	}
	actual, err := windowsDescriptorString(uintptr(unsafe.Pointer(&actualDescriptor[0])), ownerSecurityInformation|daclSecurityInformation)
	if err != nil {
		return err
	}
	expectedDescriptor, free, err := windowsDescriptor(expectedWindowsStorageSDDL(sid, directory))
	if err != nil {
		return err
	}
	defer free()
	expected, err := windowsDescriptorString(expectedDescriptor, ownerSecurityInformation|daclSecurityInformation)
	if err != nil {
		return err
	}
	if !strings.EqualFold(actual, expected) {
		return ErrUnsafePath
	}
	return nil
}

func validateWindowsStorageType(path string, directory bool) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || directory != info.IsDir() || (!directory && !info.Mode().IsRegular()) {
		return ErrUnsafePath
	}
	return nil
}

func currentWindowsSID() (string, error) {
	current, err := user.Current()
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(strings.ToUpper(current.Uid), "S-1-") {
		return "", ErrUnsafePath
	}
	return current.Uid, nil
}

func expectedWindowsStorageSDDL(sid string, directory bool) string {
	flags := ""
	if directory {
		flags = "OICI"
	}
	return "O:" + sid + "D:P(A;" + flags + ";FA;;;SY)(A;" + flags + ";FA;;;" + sid + ")"
}

func windowsStorageOwner(path string) (string, error) {
	descriptor, err := getWindowsFileDescriptor(path, ownerSecurityInformation)
	if err != nil {
		return "", err
	}
	return windowsDescriptorString(uintptr(unsafe.Pointer(&descriptor[0])), ownerSecurityInformation)
}

func getWindowsFileDescriptor(path string, information uint32) ([]byte, error) {
	pathPointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	var needed uint32
	result, _, callErr := getFileSecurityW.Call(uintptr(unsafe.Pointer(pathPointer)), uintptr(information), 0, 0, uintptr(unsafe.Pointer(&needed)))
	if result != 0 || callErr != errorInsufficientBuffer || needed == 0 {
		return nil, fmt.Errorf("query Windows ACL size: %w", callErr)
	}
	buffer := make([]byte, needed)
	result, _, callErr = getFileSecurityW.Call(uintptr(unsafe.Pointer(pathPointer)), uintptr(information), uintptr(unsafe.Pointer(&buffer[0])), uintptr(needed), uintptr(unsafe.Pointer(&needed)))
	if result == 0 {
		return nil, fmt.Errorf("read Windows ACL: %w", callErr)
	}
	return buffer, nil
}

func windowsDescriptor(sddl string) (uintptr, func(), error) {
	sddlPointer, err := syscall.UTF16PtrFromString(sddl)
	if err != nil {
		return 0, func() {}, err
	}
	var descriptor uintptr
	var size uint32
	result, _, callErr := convertStringSecurityDescriptorToSD.Call(uintptr(unsafe.Pointer(sddlPointer)), securityDescriptorRevision, uintptr(unsafe.Pointer(&descriptor)), uintptr(unsafe.Pointer(&size)))
	if result == 0 {
		return 0, func() {}, fmt.Errorf("build Windows ACL: %w", callErr)
	}
	return descriptor, func() { _, _, _ = storageLocalFree.Call(descriptor) }, nil
}

func windowsDescriptorString(descriptor uintptr, information uint32) (string, error) {
	var textPointer *uint16
	var length uint32
	result, _, callErr := convertSecurityDescriptorToStringSD.Call(descriptor, securityDescriptorRevision, uintptr(information), uintptr(unsafe.Pointer(&textPointer)), uintptr(unsafe.Pointer(&length)))
	if result == 0 {
		return "", fmt.Errorf("format Windows ACL: %w", callErr)
	}
	defer storageLocalFree.Call(uintptr(unsafe.Pointer(textPointer)))
	return syscall.UTF16ToString(unsafe.Slice(textPointer, length)), nil
}
