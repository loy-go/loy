//go:build windows

package dev

import (
	"fmt"
	"os/exec"
)

func configureProcessGroup(cmd *exec.Cmd) {
	cmd.Cancel = func() error {
		if cmd.Process != nil && cmd.Process.Pid > 0 {
			return killProcessGroup(cmd.Process.Pid, false)
		}
		return nil
	}
}

func killProcessGroup(pid int, force bool) error {
	args := []string{"/T", "/PID", fmt.Sprintf("%d", pid)}
	if force {
		args = append([]string{"/F"}, args...)
	}
	killCmd := exec.Command("taskkill", args...)
	return killCmd.Run()
}
