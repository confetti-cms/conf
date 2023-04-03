package services

import (
	"fmt"
	"strings"
)

func GetPatchSinceCommit(commit, root string, path string, verbose bool) string {
    if verbose {
        println("Create patch: " + path)
    }
	patch, err := GetPatchSinceCommitE(commit, root, path)
	if err != nil {
		println(err.Error())
	}
    return patch
}

func GetPatchSinceCommitE(commit, root, file string) (string, error) {
	// Get tracked changes from git diff in patch format
	st := fmt.Sprintf("cd %s && git diff %s -- %s", root, commit, file)
	out, err := RunCommand(st)
	if err != nil {
		return "", err
	}
	// If no results; get untracked changes
	if strings.Trim(out, "\n") != "" {
		return out, err
	}
	st = fmt.Sprintf("cd %s && git diff -- /dev/null %s", root, file)
	// Unknown way err is not nil
	out, _ = RunCommand(st)
	return out, nil
}
