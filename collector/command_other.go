//go:build !windows

package collector

import "os/exec"

func commandOutput(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

func commandRun(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}
