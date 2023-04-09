package services

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cast"
)

type Status string

const rPath = `(?P<path>[-_/0-9A-z.]+)`
const rFromPath = `(?P<from_path>[-_/0-9A-z.]+)`

const (
	GitStatusUnchanged Status = "."
	GitStatusUntracked Status = "?"
	GitStatusAdded     Status = "A"
	GitStatusModified  Status = "M"
	GitStatusDeleted   Status = "D"
	GitStatusRenamed   Status = "R"
)

type GitFileChange struct {
	StagedStatus   Status
	UnstagedStatus Status
	Path           string
	FromPath       string
	Score          int
}

func ChangedFilesSinceLastCommit(dir string) []GitFileChange {
	// Get all changes from git status in plain text
	st := fmt.Sprintf("git -C %s status --porcelain=2 --untracked-files=all", dir)
	raw, err := RunCommand(st)
	if err != nil {
		log.Fatal(err)
	}
	rawStatuses := strings.Split(strings.Trim(raw, "\n"), "\n")
	changes := getOrdinaryChanges(rawStatuses)
	changes = append(changes, getUntrackedChanges(rawStatuses)...)
	changes = append(changes, getRenameOrCopyChanges(rawStatuses)...)
	return changes
}

func IgnoreHidden(changes []GitFileChange) []GitFileChange {
	result := []GitFileChange{}
	for _, change := range changes {
		if IgnoreFile(change.Path) {
			continue
		}
		result = append(result, change)
	}
	return result
}

func IgnoreFile(file string) bool {
	if file == "" || file == "/" {
		return true
	}
	if strings.Contains(file, "/.") {
		return true
	}
	if strings.HasPrefix(file, ".") {
		return true
	}
	if strings.HasSuffix(file, "swp") || strings.HasSuffix(file, "~") {
		return true
	}
	return false
}

func RemoveIfDeleted(change GitFileChange, root string) bool {
	if change.UnstagedStatus != GitStatusDeleted && change.StagedStatus != GitStatusDeleted {
		return false
	}
	_, err := os.Stat(change.Path)
	if !os.IsNotExist(err) {
		return false
	}
	file := fileWithoutRoot(change.Path, root)
	err = SendDeleteSource(file)
	if err != nil {
		panic(err)
	}
	return true
}

func fileWithoutRoot(path, root string) string {
	return strings.ReplaceAll(path, root+"/", "")
}

// https://git-scm.com/docs/git-status#_stash_information
func getOrdinaryChanges(rawStatuses []string) []GitFileChange {
	compiler := regexp.MustCompile(`^1\s(?P<X>[AMD\.])(?P<Y>[MD\.]).*\s` + rPath + `$`)

	fileChanges := []GitFileChange{}
	for _, status := range rawStatuses {
		match, found := parseByRegex(status, compiler)
		if !found {
			continue
		}
		fileChanges = append(fileChanges, GitFileChange{
			// A character field contains the staged X value described, with unchanged indicated by a ".".
			StagedStatus: Status(match["X"]),
			// A character field contains the unstaged Y value described, with unchanged indicated by a ".".
			UnstagedStatus: Status(match["Y"]),
			Path:           match["path"],
		})
	}

	return fileChanges
}

// https://git-scm.com/docs/git-status#_stash_information
func getRenameOrCopyChanges(rawStatuses []string) []GitFileChange {
	compiler := regexp.MustCompile(`^2\s(?P<X>[MR\.])(?P<Y>[MR\.]).*\sR(?P<score>\d{1,3})\s` + rPath + `\s` + rFromPath + `$`)

	fileChanges := []GitFileChange{}
	for _, status := range rawStatuses {
		match, found := parseByRegex(status, compiler)
		if !found {
			continue
		}
		fileChanges = append(fileChanges, GitFileChange{
			// A character field contains the staged X value described, with unchanged indicated by a ".".
			StagedStatus: Status(match["X"]),
			// A character field contains the unstaged Y value described, with unchanged indicated by a ".".
			UnstagedStatus: Status(match["Y"]),
			FromPath:       match["from_path"],
			Path:           match["path"],
			Score:          cast.ToInt(match["score"]),
		})
	}

	return fileChanges
}

// https://git-scm.com/docs/git-status#_stash_information
func getUntrackedChanges(rawStatuses []string) []GitFileChange {
	compiler := regexp.MustCompile(`^\?\s` + rPath + `$`)

	fileChanges := []GitFileChange{}
	for _, status := range rawStatuses {
		match, found := parseByRegex(status, compiler)
		if !found {
			continue
		}
		fileChanges = append(fileChanges, GitFileChange{
			UnstagedStatus: GitStatusUntracked,
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
