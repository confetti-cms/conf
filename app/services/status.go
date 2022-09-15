package services

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

type Status string

const (
	None            Status = ""
	StatusUntracked Status = "?"
	StatusAdded     Status = "A"
)

type FileChange struct {
	StagedStatus   Status
	UnstagedStatus Status
	Path           string
}

func ChangedFiles(dir string) []FileChange {
	// Get all changes from git status in plain text
	st := fmt.Sprintf("git -C %s status --porcelain=2 --untracked-files=all", dir)
	raw, err := runCommand(st)
	if err != nil {
		log.Fatal(err)
	}
	rawStatuses := strings.Split(raw, "\n")

	return append(getOrdinaryChanges(rawStatuses), getUntrackedChanges(rawStatuses)...)
}

// https://git-scm.com/docs/git-status#_stash_information
func getOrdinaryChanges(rawStatuses []string) []FileChange {
	compiler := regexp.MustCompile(`^1\s(?P<X>[a.]).*\s(?P<path>[/A-z.]+)$`)

	fileChanges := []FileChange{}
	for _, status := range rawStatuses {
		match, found := parseByRegex(status, compiler)
		if !found {
			continue
		}
		fileChanges = append(fileChanges, FileChange{
			// A character field contains the staged X value described, with unchanged indicated by a ".".
			StagedStatus: Status(match["X"]),
			// A character field contains the unstaged Y value described, with unchanged indicated by a ".".
			UnstagedStatus: Status(""),
			Path:           match["path"],
		})
	}

	return fileChanges
}

// https://git-scm.com/docs/git-status#_stash_information
func getUntrackedChanges(rawStatuses []string) []FileChange {
	compiler := regexp.MustCompile(`^\?\s(?P<path>[/A-z.]+)$`)

	fileChanges := []FileChange{}
	for _, status := range rawStatuses {
		match, found := parseByRegex(status, compiler)
		if !found {
			continue
		}
		fileChanges = append(fileChanges, FileChange{
			UnstagedStatus: StatusUntracked,
			Path:           match["path"],
		})
	}

	return fileChanges
}

func parseByRegex(content string, compiler *regexp.Regexp) (map[string]string, bool) {
	mapped := map[string]string{}
	match := compiler.FindStringSubmatch(content)
	// If none found, return false
	if len(match) == 0 {
		return mapped, false
	}
	names := subExpNames(compiler)
	for iNames, name := range names {
		if len(match) <= iNames {
			continue
		}
		mapped[name] = match[iNames+1]
	}
	return mapped, true
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
