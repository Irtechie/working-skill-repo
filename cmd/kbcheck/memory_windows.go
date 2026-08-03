//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

// memoryStatusEx mirrors the Win32 MEMORYSTATUSEX structure.
type memoryStatusEx struct {
	length               uint32
	memoryLoad           uint32
	totalPhys            uint64
	availPhys            uint64
	totalPageFile        uint64
	availPageFile        uint64
	totalVirtual         uint64
	availVirtual         uint64
	availExtendedVirtual uint64
}

// availableProcessMemoryBytes reports the commit charge still available to new
// processes. Windows refuses a fork once commit is exhausted, which is the
// limit that actually bounds how many git subprocesses the suites can spawn --
// physical free RAM is not the binding constraint.
func availableProcessMemoryBytes() (uint64, bool) {
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx")
	if err := proc.Find(); err != nil {
		return 0, false
	}
	status := memoryStatusEx{}
	status.length = uint32(unsafe.Sizeof(status))
	ret, _, _ := proc.Call(uintptr(unsafe.Pointer(&status)))
	if ret == 0 || status.availPageFile == 0 {
		return 0, false
	}
	return status.availPageFile, true
}
