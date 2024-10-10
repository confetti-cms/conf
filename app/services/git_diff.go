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
	st = fmt.Sprintf("cd %s && git diff -- /dev/null %s -- %s", config.Path.Root, binaryFlag, file)
	// Unknown way err is not nil
	out, _ = RunCommand(st)
	return out, nil
}

func isBinaryFile(file string) bool {
	// This needs to be the same list as the list in the parser service
	textExtensions := []string{".txt", ".md", ".json", ".xml", ".html", ".css", ".js", ".go", ".java", ".py", ".rb", ".php", ".c", ".cpp", ".h", ".hpp", ".cs", ".ts", ".sql", ".sh", ".bat", ".ps1", ".psm1", ".psd1", ".ps1xml", ".pssc", ".psc1", ".php", ".phtml", ".inc", ".tpl", ".twig"}
	for _, ext := range textExtensions {
		if strings.HasSuffix(file, ext) {
			return false
		}
	}
	return true
}
