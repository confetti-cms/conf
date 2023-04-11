package services

import (
	"os/exec"
)

func RunCommand(command string) (string, error) {
	println(command)
	out, err := exec.Command("/bin/sh", "-c", command).Output()
	println(string(out))
	return string(out), err
}
