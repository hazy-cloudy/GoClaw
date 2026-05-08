//go:build windows

package tools

import (
	"os/exec"
	"strconv"

	"github.com/sipeed/picoclaw/pkg/processutil"
)

func prepareCommandForTermination(cmd *exec.Cmd) {
	processutil.PrepareBackgroundCommand(cmd)
}

func terminateProcessTree(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	pid := cmd.Process.Pid
	if pid <= 0 {
		return nil
	}

	taskkill := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(pid))
	processutil.PrepareBackgroundCommand(taskkill)
	_ = taskkill.Run()
	_ = cmd.Process.Kill()
	return nil
}
