package services

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/spf13/cast"
)

func GitAdd(path string) (string, error) {
	return RunCommand(fmt.Sprintf(`git add -A "%s"`, path))
}

func GitCommit(path string) (string, error) {
	return RunCommand(fmt.Sprintf(`git commit -m "Message" "%s"`, path))
}

func GitRemoteCommit(dir string) string {
	raw, err := RunCommand("cd " + dir + " && git for-each-ref refs/remotes/origin --count 1 --format \"%(objectname)\"")
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			log.Fatal(cast.ToString(ee.Stderr))
		}
		log.Fatal(err)
	}
	if strings.Contains(raw, "fatal") {
		log.Fatal(raw)
	}
	return strings.Trim(raw, "\n")
}
