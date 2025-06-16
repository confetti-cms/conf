package services

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"src/config"
)

func RunCommand(command string) (string, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/C", command)
	default:
		cmd = exec.Command("/bin/sh", "-c", command)
	}

	out, err := cmd.CombinedOutput()

	debugCommand(command, out)

	if err != nil {
		return string(out), fmt.Errorf("%v.\n\nOutput:\n\n%s", err, string(out))
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

func StreamCommand(command string) (string, error) {
	var cmd *exec.Cmd
	var out bytes.Buffer

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/C", command)
	default:
		cmd = exec.Command("/bin/sh", "-c", command)
	}

	cmd.Stdout = io.MultiWriter(os.Stdout, &out)

	var stderr bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

	if err := cmd.Start(); err != nil {
		return out.String(), fmt.Errorf("failed to start command: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		if stderr.Len() > 0 {
			return out.String(), fmt.Errorf("command execution error: %v, stderr: %s", err, stderr.String())
		}
		return out.String(), fmt.Errorf("command execution error: %v", err)
	}

	if stderr.Len() > 0 {
		return out.String(), fmt.Errorf("stderr: %s", stderr.String())
	}

	return out.String(), nil
}
