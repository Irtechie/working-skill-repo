//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"
)

const (
	amrJobObjectExtendedLimitInformationClass = 9
	amrJobObjectLimitKillOnJobClose           = 0x00002000
	amrCreateSuspended                        = 0x00000004
	amrProcessSetQuota                        = 0x0100
	amrProcessTerminate                       = 0x0001
	amrProcessSuspendResume                   = 0x0800
)

var (
	amrKernel32                 = syscall.NewLazyDLL("kernel32.dll")
	amrNtdll                    = syscall.NewLazyDLL("ntdll.dll")
	amrCreateJobObjectW         = amrKernel32.NewProc("CreateJobObjectW")
	amrSetInformationJobObject  = amrKernel32.NewProc("SetInformationJobObject")
	amrAssignProcessToJobObject = amrKernel32.NewProc("AssignProcessToJobObject")
	amrTerminateJobObject       = amrKernel32.NewProc("TerminateJobObject")
	amrCloseHandle              = amrKernel32.NewProc("CloseHandle")
	amrOpenProcess              = amrKernel32.NewProc("OpenProcess")
	amrNtResumeProcess          = amrNtdll.NewProc("NtResumeProcess")
)

type amrJobObjectBasicLimitInformation struct {
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

type amrIOCounters struct {
	ReadOperationCount, WriteOperationCount, OtherOperationCount uint64
	ReadTransferCount, WriteTransferCount, OtherTransferCount    uint64
}

type amrJobObjectExtendedLimitInformation struct {
	BasicLimitInformation amrJobObjectBasicLimitInformation
	IoInfo                amrIOCounters
	ProcessMemoryLimit    uintptr
	JobMemoryLimit        uintptr
	PeakProcessMemoryUsed uintptr
	PeakJobMemoryUsed     uintptr
}

type processTreeHandle struct {
	job syscall.Handle
}

func configureProcessTree(cmd *exec.Cmd) error {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags |= amrCreateSuspended
	return nil
}

func attachProcessTree(cmd *exec.Cmd) (*processTreeHandle, error) {
	if cmd.Process == nil {
		return nil, fmt.Errorf("process was not started")
	}
	jobRaw, _, callErr := amrCreateJobObjectW.Call(0, 0)
	if jobRaw == 0 {
		return nil, fmt.Errorf("create job object: %w", callErr)
	}
	job := syscall.Handle(jobRaw)
	info := amrJobObjectExtendedLimitInformation{}
	info.BasicLimitInformation.LimitFlags = amrJobObjectLimitKillOnJobClose
	result, _, callErr := amrSetInformationJobObject.Call(uintptr(job), amrJobObjectExtendedLimitInformationClass, uintptr(unsafe.Pointer(&info)), unsafe.Sizeof(info))
	if result == 0 {
		_, _, _ = amrCloseHandle.Call(uintptr(job))
		return nil, fmt.Errorf("configure job object: %w", callErr)
	}
	processRaw, _, callErr := amrOpenProcess.Call(amrProcessSetQuota|amrProcessTerminate|amrProcessSuspendResume, 0, uintptr(uint32(cmd.Process.Pid)))
	if processRaw == 0 {
		_, _, _ = amrCloseHandle.Call(uintptr(job))
		return nil, fmt.Errorf("open process: %w", callErr)
	}
	result, _, callErr = amrAssignProcessToJobObject.Call(uintptr(job), processRaw)
	if result == 0 {
		_, _, _ = amrCloseHandle.Call(processRaw)
		_, _, _ = amrCloseHandle.Call(uintptr(job))
		return nil, fmt.Errorf("assign job object: %w", callErr)
	}
	status, _, callErr := amrNtResumeProcess.Call(processRaw)
	_, _, _ = amrCloseHandle.Call(processRaw)
	if status != 0 {
		_, _, _ = amrTerminateJobObject.Call(uintptr(job), 1)
		_, _, _ = amrCloseHandle.Call(uintptr(job))
		return nil, fmt.Errorf("resume process: status=0x%x err=%w", status, callErr)
	}
	return &processTreeHandle{job: job}, nil
}

func (h *processTreeHandle) Kill() error {
	if h == nil || h.job == 0 {
		return nil
	}
	result, _, callErr := amrTerminateJobObject.Call(uintptr(h.job), 1)
	if result == 0 {
		return fmt.Errorf("terminate job: %w", callErr)
	}
	return nil
}

func (h *processTreeHandle) Close() error {
	if h == nil || h.job == 0 {
		return nil
	}
	handle := h.job
	h.job = 0
	result, _, callErr := amrCloseHandle.Call(uintptr(handle))
	if result == 0 {
		return fmt.Errorf("close job: %w", callErr)
	}
	return nil
}
