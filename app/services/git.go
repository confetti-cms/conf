package services

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cast"
)

func GitAdd(path string) (string, error) {
	return runCommand(fmt.Sprintf(`git add "%s"`, path))
}

func runCommand(command string) (string, error) {
	out, err := exec.Command("/bin/sh", "-c", command).Output()
	return cast.ToString(out), err
}
