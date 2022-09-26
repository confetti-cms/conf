package services

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cast"
)

func GetPatchSinceOrigin(file string) (string, error) {
	// Get all changes from git diff in patch format
	st := fmt.Sprintf("git diff %s", file)
	raw, err := RunCommand(st)
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			// Use the first line of execution error
			message := cast.ToString(ee.Stderr)
			firstLine := strings.Split(message, "\n")[0]
			return "", errors.New(firstLine)
		}
	}
	return raw, err
}
