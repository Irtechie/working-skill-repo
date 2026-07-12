//go:build !windows

package main

import (
	"fmt"
	"os/exec"
	"syscall"
)

type processTreeHandle struct {
	pgid int
}

func configureProcessTree(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return nil
}

func attachProcessTree(cmd *exec.Cmd) (*processTreeHandle, error) {
	if cmd.Process == nil {
		return nil, fmt.Errorf("process was not started")
	}
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		return nil, err
	}
	return &processTreeHandle{pgid: pgid}, nil
}

func (h *processTreeHandle) Kill() error {
	if h == nil || h.pgid == 0 {
		return nil
	}
	return syscall.Kill(-h.pgid, syscall.SIGKILL)
}

func (h *processTreeHandle) Close() error { return nil }
