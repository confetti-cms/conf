package services

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/spf13/cast"
)

func GetPatchSinceCommit(commit, file string) (string, error) {
	// Get all changes from git diff in patch format
	st := fmt.Sprintf("git diff %s -- %s", commit, file)
	raw, err := RunCommand(st)
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", errors.New(cast.ToString(ee.Stderr))
		}
	}
	return raw, err
}
