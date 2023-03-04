package services

import (
	"fmt"
	"path/filepath"
	"strings"
)

func SendPatchSinceCommit(commit, root string, path string) {
	println("Create patch: " + path)
	// Ignore hidden files and directories
	if strings.HasPrefix(path, ".") || strings.HasPrefix(filepath.Base(path), ".") {
		return
	}
	patch, err := GetPatchSinceCommit(commit, root, path)
	if err != nil {
		println(err.Error())
	}
	if patch == "" {
		println("Ignore (no change in patch): " + path)
		return
	}
	println("Send patch: " + path)
	err = SendPatch(PatchBody{
		Path:  path,
		Patch: patch,
	})
	if err != nil {
		println("Err:")
		println(err.Error())
	}
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
