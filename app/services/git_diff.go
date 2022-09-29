package services

import (
	"fmt"
	"log"
	"strings"
)

func SendPatchSinceCommit(commit, root string, path string) {
	patch, err := GetPatchSinceCommit(commit, root, path)
	if err != nil {
		log.Fatal(err)
	}
	err = SendPatch(PatchBody{
		Path:  path,
		Patch: patch,
	})
	if err != nil {
		log.Fatal(err)
	}
	println("Has send: " + path)
}

func GetPatchSinceCommit(commit, root, path string) (string, error) {
	// Get tracked changes from git diff in patch format
	st := fmt.Sprintf("cd %s && git diff %s -- %s", root, commit, path)
	out, err := RunCommand(st)
	if err != nil {
		return "", err
	}
	// If no results; get untracked changes
	if strings.Trim(out, "\n") != "" {
		return out, err
	}
	st = fmt.Sprintf("cd %s && git diff -- /dev/null %s", root, path)
	// Unknown way err is not nil
	out, _ = RunCommand(st)
	return out, nil
}
