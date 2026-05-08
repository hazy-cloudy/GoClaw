//go:build !windows

package processutil

import "os/exec"

// PrepareBackgroundCommand is a no-op on non-Windows platforms.
func PrepareBackgroundCommand(cmd *exec.Cmd) {
}
