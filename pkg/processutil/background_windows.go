//go:build windows

package processutil

import (
	"os/exec"
	"syscall"
)

const createNoWindow = 0x08000000

// PrepareBackgroundCommand ensures Windows subprocesses started from GUI flows
// stay silent instead of creating visible console windows.
func PrepareBackgroundCommand(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}

	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= createNoWindow
}
