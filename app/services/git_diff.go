package services

import (
	"fmt"
	"src/config"
	"strings"
)

var ErrNewFileEmptyPatch = fmt.Errorf("Patch is empty. This may be due to the editor. They may have created a new file and edit it directly. So the first patch is empty.")

func GetPatchSinceCommit(commit, path string, isNew bool) (string, error) {
	if config.App.VeryVerbose {
		println("Create patch: " + path)
	}
	return GetPatchSinceCommitE(commit, path, isNew)
}

func GetPatchSinceCommitE(commit, file string, isNew bool) (string, error) {
	// Determine if the file is binary and set the binary flag
	binaryFlag := ""
	if isBinaryFile(file) {
		binaryFlag = " --binary"
	}

	// Get tracked changes from git diff in patch format
	st := fmt.Sprintf("cd %s && git diff %s%s -- %s", config.Path.Root, commit, binaryFlag, file)
	out, err := RunCommand(st)
	if err != nil {
		return "", err
	}
	if strings.Trim(out, "\n") != "" || !isNew {
		return out, err
	}

	// If no results; get untracked changes
	st = fmt.Sprintf("cd %s && git diff %s -- /dev/null %s", config.Path.Root, binaryFlag, file)
	out, _ = RunCommand(st)
	if out == "" {
		return "", ErrNewFileEmptyPatch
	}
	return out, nil
}

func isBinaryFile(file string) bool {
	// This needs to be the same list as the list in the parser service
	textExtensions := []string{".txt", ".md", ".json", ".xml", ".html", ".css", ".js", ".go", ".java", ".py", ".rb", ".php", ".c", ".cpp", ".h", ".hpp", ".cs", ".ts", ".sql", ".sh", ".bat", ".ps1", ".psm1", ".psd1", ".ps1xml", ".pssc", ".psc1", ".phtml", ".inc", ".tpl", ".twig"}
	for _, ext := range textExtensions {
		if strings.HasSuffix(file, ext) {
			return false
		}
	}
	return true
}
