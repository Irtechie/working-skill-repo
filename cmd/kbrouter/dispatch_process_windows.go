//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"
)

const (
	jobObjectExtendedLimitInformationClass = 9
	jobObjectLimitKillOnJobClose           = 0x00002000
	createSuspended                        = 0x00000004
	processSetQuota                        = 0x0100
	processTerminate                       = 0x0001
	processSuspendResume                   = 0x0800
)

var (
	processKernel32          = syscall.NewLazyDLL("kernel32.dll")
	processNtdll             = syscall.NewLazyDLL("ntdll.dll")
	createJobObjectW         = processKernel32.NewProc("CreateJobObjectW")
	setInformationJobObject  = processKernel32.NewProc("SetInformationJobObject")
	assignProcessToJobObject = processKernel32.NewProc("AssignProcessToJobObject")
	terminateJobObject       = processKernel32.NewProc("TerminateJobObject")
	closeHandle              = processKernel32.NewProc("CloseHandle")
	openProcess              = processKernel32.NewProc("OpenProcess")
	ntResumeProcess          = processNtdll.NewProc("NtResumeProcess")
)

type jobObjectBasicLimitInformation struct {
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

type ioCounters struct {
	ReadOperationCount  uint64
	WriteOperationCount uint64
	OtherOperationCount uint64
	ReadTransferCount   uint64
	WriteTransferCount  uint64
	OtherTransferCount  uint64
}

type jobObjectExtendedLimitInformation struct {
	BasicLimitInformation jobObjectBasicLimitInformation
	IoInfo                ioCounters
	ProcessMemoryLimit    uintptr
	JobMemoryLimit        uintptr
	PeakProcessMemoryUsed uintptr
	PeakJobMemoryUsed     uintptr
}

type processTreeHandle struct {
	job syscall.Handle
}

func ensureProcessTreeContainment() error {
	return nil
}

func configureProcessTree(cmd *exec.Cmd) error {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags |= createSuspended
	return nil
}

func attachProcessTree(cmd *exec.Cmd) (*processTreeHandle, error) {
	if cmd.Process == nil {
		return nil, fmt.Errorf("process was not started")
	}
	jobRaw, _, err := createJobObjectW.Call(0, 0)
	if jobRaw == 0 {
		return nil, fmt.Errorf("create Windows job object: %w", err)
	}
	job := syscall.Handle(jobRaw)
	info := jobObjectExtendedLimitInformation{}
	info.BasicLimitInformation.LimitFlags = jobObjectLimitKillOnJobClose
	result, _, err := setInformationJobObject.Call(
		uintptr(job),
		uintptr(jobObjectExtendedLimitInformationClass),
		uintptr(unsafe.Pointer(&info)),
		unsafe.Sizeof(info),
	)
	if result == 0 {
		_, _, _ = closeHandle.Call(uintptr(job))
		return nil, fmt.Errorf("configure Windows job object: %w", err)
	}
	processRaw, _, err := openProcess.Call(uintptr(processSetQuota|processTerminate|processSuspendResume), 0, uintptr(uint32(cmd.Process.Pid)))
	if processRaw == 0 {
		_, _, _ = closeHandle.Call(uintptr(job))
		return nil, fmt.Errorf("open child for Windows job assignment: %w", err)
	}
	result, _, err = assignProcessToJobObject.Call(uintptr(job), processRaw)
	if result == 0 {
		_, _, _ = closeHandle.Call(processRaw)
		_, _, _ = closeHandle.Call(uintptr(job))
		return nil, fmt.Errorf("assign child to Windows job object: %w", err)
	}
	if err := resumeWindowsProcess(processRaw); err != nil {
		_, _, _ = closeHandle.Call(processRaw)
		_, _, _ = terminateJobObject.Call(uintptr(job), 1)
		_, _, _ = closeHandle.Call(uintptr(job))
		return nil, err
	}
	_, _, _ = closeHandle.Call(processRaw)
	return &processTreeHandle{job: job}, nil
}

func resumeWindowsProcess(process uintptr) error {
	status, _, err := ntResumeProcess.Call(process)
	if status != 0 {
		return fmt.Errorf("resume Windows process: status=0x%x err=%w", status, err)
	}
	return nil
}

func (h *processTreeHandle) Kill() error {
	if h == nil || h.job == 0 {
		return nil
	}
	result, _, err := terminateJobObject.Call(uintptr(h.job), 1)
	if result == 0 {
		return fmt.Errorf("terminate Windows job object: %w", err)
	}
	return nil
}

func (h *processTreeHandle) Close() error {
	if h == nil || h.job == 0 {
		return nil
	}
	handle := h.job
	result, _, err := closeHandle.Call(uintptr(handle))
	if result == 0 {
		return fmt.Errorf("close Windows job object: %w", err)
	}
	h.job = 0
	return nil
}
