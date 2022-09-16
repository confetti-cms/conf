package services

import (
	"os/exec"

	"github.com/spf13/cast"
)

func RunCommand(command string) (string, error) {
	out, err := exec.Command("/bin/sh", "-c", command).Output()
	return cast.ToString(out), err
}
