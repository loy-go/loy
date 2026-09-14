//go:build !windows

package dev

import (
	"os/exec"
	"syscall"
)

func configureProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.Cancel = func() error {
		if cmd.Process != nil && cmd.Process.Pid > 1 {
			return killProcessGroup(cmd.Process.Pid, false)
		}
		return nil
	}
}

func killProcessGroup(pid int, force bool) error {
	pgid, err := syscall.Getpgid(pid)
	if err != nil || pgid <= 1 {
		pgid = pid
	}
	sig := syscall.SIGTERM
	if force {
		sig = syscall.SIGKILL
	}
	return syscall.Kill(-pgid, sig)
}
