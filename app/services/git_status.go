package services

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"src/config"
	"strings"

	"github.com/confetti-framework/framework/inter"
)

type Status string

const rPath = `(?P<path>[-_\\/0-9A-z.]+)`
const rToPath = `(?P<to_path>[-_\\/0-9A-z.]*)`

const (
	GitStatusUnchanged  Status = "."
	GitStatusUntracked  Status = "?"
	GitStatusAdded      Status = "A"
	GitStatusCopy       Status = "C"
	GitStatusDeleted    Status = "D"
	GitStatusModified   Status = "M"
	GitStatusRenamed    Status = "R"
	GitStatusChangeType Status = "T"
	GitStatusUnmerged   Status = "U"
)

type GitFileChange struct {
	Status Status
	Path   string
}

func ChangedFilesSinceRemoteCommit(remoteCommit string) []GitFileChange {
	changes := []GitFileChange{}
	// Get all changes from git status in plain text
	cm := fmt.Sprintf("cd %s && git diff %s --name-status", config.Path.Root, remoteCommit)
	raw, err := RunCommand(cm)
	if err != nil {
		println("Err: from command: " + cm)
		log.Fatal(err)
	}
	rawStatuses := strings.Split(strings.Trim(raw, "\n"), "\n")
	// Staged files
	if remoteCommit == "" {
		cm = fmt.Sprintf("cd %s && git diff --name-status --staged", config.Path.Root)
		raw, err = RunCommand(cm)
		if err != nil {
			println("Err: from command: " + cm)
			log.Fatal(err)
		}
		result := strings.Split(strings.Trim(raw, "\n"), "\n")
		rawStatuses = append(rawStatuses, result...)
	}
	changes = getOrdinaryChanges(rawStatuses)
	// Get all untracked (new) files
	cm = fmt.Sprintf("cd %s && git ls-files --others --exclude-standard", config.Path.Root)
	raw, err = RunCommand(cm)
	if err != nil {
		println("Err: from command: " + cm)
		log.Fatal(err)
	}
	rawStatuses = strings.Split(strings.Trim(raw, "\n"), "\n")
	changes = append(changes, getChangesByList(rawStatuses)...)

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
	if file == "" || file == config.App.LineSeparator {
		if config.App.VeryVeryVerbose {
			fmt.Printf("Directory %s is root so we ignore it\n", file)
		}
		return true
	}
	if strings.Contains(file, config.App.LineSeparator+".") {
		if config.App.VeryVeryVerbose {
			fmt.Printf("File or directory %s is hidden so we ignore it\n", file)
		}
		return true
	}
	if strings.HasPrefix(file, ".") {
		if config.App.VeryVeryVerbose {
			fmt.Printf("File or directory %s is hidden so we ignore it\n", file)
		}
		return true
	}
	if strings.HasSuffix(file, "swp") || strings.HasSuffix(file, "~") {
		if config.App.VeryVerbose {
			fmt.Printf("Ignoring file %s because it appears to be an unsaved file (a temporary swap file created during editing).\n", file)
		}
		return true
	}
	return false
}

func RemoveIfDeleted(cli inter.Cli, env Environment, change GitFileChange, repo string) bool {
	if change.Status != GitStatusDeleted {
		return false
	}
	_, err := os.Stat(change.Path)
	if !os.IsNotExist(err) {
		return false
	}
	file := fileWithoutRoot(change.Path, config.Path.Root)
	err = SendDeleteSource(cli, env, file, repo)
	if err != nil {
		cli.Error(err.Error())
		if !errors.Is(err, UserError) {
			log.Fatal(err)
		}
	}
	return true
}

func fileWithoutRoot(path, root string) string {
	return strings.ReplaceAll(path, root, "")
}

// https://git-scm.com/docs/git-status#_stash_information
func getOrdinaryChanges(rawStatuses []string) []GitFileChange {
	fileChanges := []GitFileChange{}
	compiler := regexp.MustCompile(`^(?P<status>[ACDMTUXR\.])[\s\t\d]+` + rPath + `\s*` + rToPath + `$`)
	for _, rawStatus := range rawStatuses {
		match, found := parseByRegex(rawStatus, compiler)
		if !found {
			continue
		}
		status := Status(match["status"])
		if status == GitStatusRenamed {
			fileChanges = append(fileChanges, GitFileChange{
				Status: GitStatusDeleted,
				Path:   match["path"],
			})
			fileChanges = append(fileChanges, GitFileChange{
				Status: status,
				Path:   match["to_path"],
			})
			continue
		}
		fileChanges = append(fileChanges, GitFileChange{
			Status: status,
			Path:   match["path"],
		})
	}

	return fileChanges
}

// https://git-scm.com/docs/git-status#_stash_information
func getChangesByList(files []string) []GitFileChange {
	fileChanges := []GitFileChange{}
	for _, file := range files {
		if file == "" {
			continue
		}
		fileChanges = append(fileChanges, GitFileChange{
			Status: GitStatusAdded,
			Path:   file,
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
