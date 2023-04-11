package services

import (
	"os/exec"
)

func RunCommand(command string) (string, error) {
	out, err := exec.Command("/bin/sh", "-c", command).Output()
	return string(out), err
}
