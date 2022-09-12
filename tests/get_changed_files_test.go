package test

import (
	"log"
	"os"
	"os/exec"
	"src/app/services"
	"testing"

	"github.com/matryer/is"
)

const mockDir = "mock_generated"

func Test_no_changed_files(t *testing.T) {
	// Given
	dir := generateGitDir("no_changed_files")
	// When
	changes := services.ChangedFiles(dir)
	// Then
	i := is.New(t)
	i.True(len(changes) == 0)
}

func Test_one_changed_file(t *testing.T) {
	// Given
	dir := generateGitDir("one_changed_file")
	touchFile(dir, "logo.svg")
	// When
	changes := services.ChangedFiles(dir)
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal("logo.svg", changes[0].Path)
}

func generateGitDir(testDir string) string {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	dir := currentDir + "/" + mockDir + "/" + testDir
	// Clean up directory from old test
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		err := os.RemoveAll(dir)
		if err != nil {
			log.Fatal(err)
		}
	}
	// Create directory
	if err := os.Mkdir(dir, 0755); err != nil {
		log.Fatal(err)
	}
	// Create empty git repository
	cmd := exec.Command("git", "init", dir)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	return dir
}

func touchFile(dir string, name string) {
	file, err := os.Create(dir + "/" + name)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
}
