package main

import (
	"os/exec"
	"syscall"
)

// hideWindow keeps the console window of a helper program (PowerShell) from flashing on the screen.
func hideWindow(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
