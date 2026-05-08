//go:build windows

package tools

import (
	"os/exec"
	"strconv"

	"github.com/sipeed/picoclaw/pkg/processutil"
)

func killProcessGroup(pid int) error {
	cmd := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(pid))
	processutil.PrepareBackgroundCommand(cmd)
	_ = cmd.Run()
	return nil
}
