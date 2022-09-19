package services

import (
	"fmt"
)

func GitAdd(path string) (string, error) {
	return RunCommand(fmt.Sprintf(`git add -A "%s"`, path))
}

func GitCommit(path string) (string, error) {
	return RunCommand(fmt.Sprintf(`git commit -m "Message" "%s"`, path))
}
