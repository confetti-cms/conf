package services

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"src/config"
	"strings"

	"github.com/spf13/cast"
)

func GetRepositoryName(root string) (string, error) {
	// output example: git@github.com:confetti-cms/office.git
	// output example: https://github.com/confetti-cms/office.git
	output, err := RunCommand(fmt.Sprintf(`cd %s && git config --get remote.origin.url`, root))
	if err != nil {
		return "", fmt.Errorf("failed to get repository name: %v", err)
	}
	output = strings.TrimSpace(output)
	result, err := GetRepositoryNameByOriginUrl(output)
	if config.App.Debug {
		fmt.Printf("Current repository: %s", result)
	}
	return result, err
}

func GetRepositoryNameByOriginUrl(url string) (string, error) {
	// url example: git@github.com:confetti-cms/office.git
	// url example: https://github.com/.git
	re := regexp.MustCompile(`([^/:]*/[^/]*)\.git$`)
	match := re.FindStringSubmatch(url)
	if len(match) != 2 {
		return "", errors.New("failed to parse repo name from url: '" + url + "'")
	}
	return match[1], nil
}

func GitAdd(path string) (string, error) {
	return RunCommand(fmt.Sprintf(`git add -A "%s"`, path))
}

func GitCommit(path string) (string, error) {
	return RunCommand(fmt.Sprintf(`git commit -m "Message" "%s"`, path))
}

func GitIgnored(root, dir string) bool {
	if dir == root {
		return false
	}
	dir = strings.TrimPrefix(dir, root)
	cmd := fmt.Sprintf(`cd %s && git check-ignore %s`, root, dir)
	out, _ := RunCommand(cmd)
	// Ignore the error (exit status 1)
	if out == "" {
		return false
	}
	return true
}

func GitRemoteCommit(dir string) string {
	cmd := "cd " + dir + " && git for-each-ref refs/remotes/origin --count 1 --format \"%(objectname)\""
	out, err := RunCommand(cmd)
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			log.Fatal(cast.ToString(ee.Stderr))
		}
		log.Fatal(err)
	}
	if strings.Contains(out, "fatal") {
		log.Fatal(out)
	}
	return strings.Trim(out, "\n")
}
