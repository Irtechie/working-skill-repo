//go:build !windows

package main

import (
	"errors"
	"fmt"
	"os/exec"
	"syscall"
)

type checkProcessTreeHandle struct{ pid int }

func configureCheckProcessTree(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return nil
}

func attachCheckProcessTree(cmd *exec.Cmd) (*checkProcessTreeHandle, error) {
	if cmd.Process == nil {
		return nil, fmt.Errorf("process was not started")
	}
	return &checkProcessTreeHandle{pid: cmd.Process.Pid}, nil
}

func (handle *checkProcessTreeHandle) Kill() error {
	if handle == nil || handle.pid <= 0 {
		return nil
	}
	err := syscall.Kill(-handle.pid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}

func (handle *checkProcessTreeHandle) Close() error {
	if handle == nil || handle.pid <= 0 {
		return nil
	}
	err := handle.Kill()
	handle.pid = 0
	return err
}
