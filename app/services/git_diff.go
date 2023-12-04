package services

import (
	"fmt"
	"src/config"
	"strings"
)

func GetPatchSinceCommit(commit, path string, isNew bool) string {
	if config.App.Debug {
		println("Create patch: " + path)
	}
	patch, err := GetPatchSinceCommitE(commit, path, isNew)
	if err != nil {
		println(err.Error())
	}
	return patch
}

func GetPatchSinceCommitE(commit, file string, isNew bool) (string, error) {
	// Get tracked changes from git diff in patch format
	st := fmt.Sprintf("cd %s && git diff %s -- %s", config.Path.Root, commit, file)
	out, err := RunCommand(st)
	if err != nil {
		return "", err
	}
	if strings.Trim(out, "\n") != "" || isNew == false {
		return out, err
	}
	// If no results; get untracked changes
	st = fmt.Sprintf("cd %s && git diff -- /dev/null %s", config.Path.Root, file)
	// Unknown way err is not nil
	out, _ = RunCommand(st)
	return out, nil
}
