package services

import (
	"fmt"
)

func GitAdd(path string) (string, error) {
	return RunCommand(fmt.Sprintf(`git add "%s"`, path))
}

func GitCommit(path string) (string, error) {
	return RunCommand(fmt.Sprintf(`git commit -m "Message" "%s"`, path))
}
