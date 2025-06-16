package commands

import (
	"os"
	"path/filepath"
	"src/config"
	"strings"

	"github.com/confetti-framework/errors"
)

func getDirectoryOrCurrent(dir string) (string, error) {
	if dir != "" {
		if _, err := os.Stat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
			return "", errors.New("The specified directory is incorrect. Please ensure that the given directory is correct.")
		}
		return formatRootDir(dir), nil
	}
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(path, ".git")); os.IsNotExist(err) {
		return "", errors.New("You are not running this command in the correct location. Please ensure that you are running the command in the correct Git repository.")
	}
	return formatRootDir(path), nil
}

func formatRootDir(dir string) string {
	return strings.TrimRight(dir, config.App.LineSeparator) + config.App.LineSeparator
}
