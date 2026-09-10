//go:build windows

package collector

import (
	"os/exec"
	"syscall"
)

const createNoWindow = 0x08000000

func hiddenCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	return cmd
}

func commandOutput(name string, args ...string) ([]byte, error) {
	return hiddenCommand(name, args...).Output()
}

func commandRun(name string, args ...string) error {
	return hiddenCommand(name, args...).Run()
}
