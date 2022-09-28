package services

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/spf13/cast"
)

func GetPatchSinceCommit(commit, path string) (string, error) {
	// Get all changes from git diff in patch format
	st := fmt.Sprintf("git diff %s -- %s", commit, path)
	println(st)
	raw, err := RunCommand(st)
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", errors.New(cast.ToString(ee.Stderr))
		}
	}
	return raw, err
}
