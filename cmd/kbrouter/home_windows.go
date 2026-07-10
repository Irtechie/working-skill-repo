//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

const csidlProfile = 0x0028

var shell32 = syscall.NewLazyDLL("shell32.dll")
var shGetFolderPath = shell32.NewProc("SHGetFolderPathW")

func operatingSystemUserHome() (string, error) {
	buffer := make([]uint16, syscall.MAX_PATH)
	result, _, callErr := shGetFolderPath.Call(
		0,
		uintptr(csidlProfile),
		0,
		0,
		uintptr(unsafe.Pointer(&buffer[0])),
	)
	if result != 0 {
		return "", fmt.Errorf("resolve Windows profile directory: HRESULT %#x: %v", result, callErr)
	}
	home := syscall.UTF16ToString(buffer)
	if home == "" {
		return "", fmt.Errorf("resolve Windows profile directory: empty path")
	}
	return home, nil
}
