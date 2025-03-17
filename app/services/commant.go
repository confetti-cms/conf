package services

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"src/config"
)

func RunCommand(command string) (string, error) {
	out, err := exec.Command("/bin/sh", "-c", command).CombinedOutput()

	debugCommand(command, out)

	if err != nil {
		return string(out), fmt.Errorf("%v.'\n\nOutput:\n\n%s", err, string(out))
	}
	return string(out), nil
}

func debugCommand(command string, outR []byte) {
	if config.App.VeryVeryVerbose {
		out := string(outR)
		if len(out) > 400 {
			out = out[:400] + "(...)"
		}
		fmt.Printf("Command: %s\nOutput: %s\n", command, out)
	}
}

func StreamCommand(command string) error {
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Stdout = os.Stdout

	var stderr bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("command execution error: %v, stderr: %s", err, stderr.String())
		}
		return fmt.Errorf("command execution error: %v", err)
	}

	if stderr.Len() > 0 {
		return fmt.Errorf("stderr: %s", stderr.String())
	}

	return nil
}
