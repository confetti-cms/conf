package services

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"

	"github.com/spf13/cast"
)

type Status string

const (
	StatusUntracked Status = "?"
)

type FileChange struct {
	Path   string
	Status Status
}

func ChangedFiles(dir string) []FileChange {
	// Get all changes from git status in plain text
	st := fmt.Sprintf("git -C %s status -s --porcelain --untracked-files=all", dir)
	stdout, err := runCommand(st)
	if err != nil {
		log.Fatal(err)
	}
	// Get all changes in FileChange objects
	compiler := regexp.MustCompile(`(?P<path>[/A-z.]+)`)
	matches := parseByRegex(stdout, compiler)
	fileChanges := []FileChange{}
	for _, match := range matches {
		fileChanges = append(fileChanges, FileChange{
			Path:   match["path"],
			Status: StatusUntracked,
		})
	}
	return fileChanges
}

func runCommand(command string) (string, error) {
	out, err := exec.Command("/bin/sh", "-c", command).Output()

	return cast.ToString(out), err
}

func parseByRegex(content string, compiler *regexp.Regexp) []map[string]string {
	matches := compiler.FindAllStringSubmatch(content, -1)
	result := make([]map[string]string, 0)
	names := subExpNames(compiler)
	for _, match := range matches {
		mapped := make(map[string]string)
		for iNames, name := range names {
			if len(match) <= iNames {
				continue
			}
			mapped[name] = match[iNames+1]
		}
		result = append(result, mapped)
	}

	return result
}

func subExpNames(compiler *regexp.Regexp) []string {
	names := []string{}
	for i, name := range compiler.SubexpNames() {
		if i == 0 {
			continue
		}
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	return names
}
