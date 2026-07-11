//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"
)

const (
	checkJobObjectExtendedLimitInformationClass = 9
	checkJobObjectLimitKillOnJobClose           = 0x00002000
	checkCreateSuspended                        = 0x00000004
	checkProcessSetQuota                        = 0x0100
	checkProcessTerminate                       = 0x0001
	checkProcessSuspendResume                   = 0x0800
)

var (
	checkKernel32          = syscall.NewLazyDLL("kernel32.dll")
	checkNtdll             = syscall.NewLazyDLL("ntdll.dll")
	checkCreateJobObjectW  = checkKernel32.NewProc("CreateJobObjectW")
	checkSetJobInformation = checkKernel32.NewProc("SetInformationJobObject")
	checkAssignProcess     = checkKernel32.NewProc("AssignProcessToJobObject")
	checkTerminateJob      = checkKernel32.NewProc("TerminateJobObject")
	checkCloseHandle       = checkKernel32.NewProc("CloseHandle")
	checkOpenProcess       = checkKernel32.NewProc("OpenProcess")
	checkResumeProcess     = checkNtdll.NewProc("NtResumeProcess")
)

type checkJobBasicLimitInformation struct {
	PerProcessUserTimeLimit int64
	PerJobUserTimeLimit     int64
	LimitFlags              uint32
	MinimumWorkingSetSize   uintptr
	MaximumWorkingSetSize   uintptr
	ActiveProcessLimit      uint32
	Affinity                uintptr
	PriorityClass           uint32
	SchedulingClass         uint32
}

type checkIOCounters struct {
	ReadOperationCount  uint64
	WriteOperationCount uint64
	OtherOperationCount uint64
	ReadTransferCount   uint64
	WriteTransferCount  uint64
	OtherTransferCount  uint64
}

type checkJobExtendedLimitInformation struct {
	BasicLimitInformation checkJobBasicLimitInformation
	IoInfo                checkIOCounters
	ProcessMemoryLimit    uintptr
	JobMemoryLimit        uintptr
	PeakProcessMemoryUsed uintptr
	PeakJobMemoryUsed     uintptr
}

type checkProcessTreeHandle struct{ job syscall.Handle }

func configureCheckProcessTree(cmd *exec.Cmd) error {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags |= checkCreateSuspended
	return nil
}

func attachCheckProcessTree(cmd *exec.Cmd) (*checkProcessTreeHandle, error) {
	if cmd.Process == nil {
		return nil, fmt.Errorf("process was not started")
	}
	jobRaw, _, callErr := checkCreateJobObjectW.Call(0, 0)
	if jobRaw == 0 {
		return nil, fmt.Errorf("create Windows job object: %w", callErr)
	}
	job := syscall.Handle(jobRaw)
	info := checkJobExtendedLimitInformation{}
	info.BasicLimitInformation.LimitFlags = checkJobObjectLimitKillOnJobClose
	result, _, callErr := checkSetJobInformation.Call(uintptr(job), checkJobObjectExtendedLimitInformationClass, uintptr(unsafe.Pointer(&info)), unsafe.Sizeof(info))
	if result == 0 {
		_, _, _ = checkCloseHandle.Call(uintptr(job))
		return nil, fmt.Errorf("configure Windows job object: %w", callErr)
	}
	processRaw, _, callErr := checkOpenProcess.Call(checkProcessSetQuota|checkProcessTerminate|checkProcessSuspendResume, 0, uintptr(uint32(cmd.Process.Pid)))
	if processRaw == 0 {
		_, _, _ = checkCloseHandle.Call(uintptr(job))
		return nil, fmt.Errorf("open child for Windows job assignment: %w", callErr)
	}
	result, _, callErr = checkAssignProcess.Call(uintptr(job), processRaw)
	if result == 0 {
		_, _, _ = checkCloseHandle.Call(processRaw)
		_, _, _ = checkCloseHandle.Call(uintptr(job))
		return nil, fmt.Errorf("assign child to Windows job object: %w", callErr)
	}
	status, _, callErr := checkResumeProcess.Call(processRaw)
	_, _, _ = checkCloseHandle.Call(processRaw)
	if status != 0 {
		_, _, _ = checkTerminateJob.Call(uintptr(job), 1)
		_, _, _ = checkCloseHandle.Call(uintptr(job))
		return nil, fmt.Errorf("resume Windows process: status=0x%x err=%w", status, callErr)
	}
	return &checkProcessTreeHandle{job: job}, nil
}

func (handle *checkProcessTreeHandle) Kill() error {
	if handle == nil || handle.job == 0 {
		return nil
	}
	result, _, callErr := checkTerminateJob.Call(uintptr(handle.job), 1)
	if result == 0 {
		return fmt.Errorf("terminate Windows job object: %w", callErr)
	}
	return nil
}

func (handle *checkProcessTreeHandle) Close() error {
	if handle == nil || handle.job == 0 {
		return nil
	}
	job := handle.job
	handle.job = 0
	result, _, callErr := checkCloseHandle.Call(uintptr(job))
	if result == 0 {
		return fmt.Errorf("close Windows job object: %w", callErr)
	}
	return nil
}
