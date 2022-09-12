package services

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/spf13/cast"
)

type FileChange struct {
	Path string
}

func ChangedFiles(dir string) []FileChange {
	st := fmt.Sprintf("git -C %s status -s --porcelain", dir)
	stdout, err := runCommand(st)
	if err != nil {
		log.Fatal(err)
	}
	if stdout == "" {
		return []FileChange{}
	}
	path := strings.Split(stdout, "\n")[0]
	path = strings.Split(path, " ")[1]
	file := FileChange{Path: path}
	return []FileChange{file}
}

func runCommand(command string) (string, error) {
	out, err := exec.Command("/bin/sh", "-c", command).Output()

	return cast.ToString(out), err
}
