//go:build !windows

package main

import (
	"errors"
	"fmt"
	"os/exec"
	"syscall"
)

type processTreeHandle struct {
	pid int
}

func ensureProcessTreeContainment() error {
	return fmt.Errorf("external worker dispatch is unavailable: this platform has no proven process-tree containment")
}

func configureProcessTree(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return nil
}

func attachProcessTree(cmd *exec.Cmd) (*processTreeHandle, error) {
	if cmd.Process == nil {
		return nil, fmt.Errorf("process was not started")
	}
	return &processTreeHandle{pid: cmd.Process.Pid}, nil
}

func (h *processTreeHandle) Kill() error {
	if h == nil || h.pid <= 0 {
		return nil
	}
	err := syscall.Kill(-h.pid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}

func (h *processTreeHandle) Close() error {
	if h == nil || h.pid <= 0 {
		return nil
	}
	err := syscall.Kill(-h.pid, syscall.SIGKILL)
	if err != nil && !errors.Is(err, syscall.ESRCH) {
		return err
	}
	h.pid = 0
	return nil
}
