package services

import (
	"fmt"
	"os/exec"
	"src/config"
)

func RunCommand(command string) (string, error) {
	out, err := exec.Command("/bin/sh", "-c", command).Output()
	debugCommand(command, out)
	return string(out), err
}

func debugCommand(command string, outR []byte) {
	if config.App.Debug {
		out := string(outR)
		if len(out) > 400 {
			out = out[:400] + "(...)"
		}
		fmt.Printf("Command: %s\nOutput: %s\n", command, out)
	}
}
