package test

import (
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"src/app/services"
	"testing"

	"github.com/matryer/is"
)

const mockDir = "mock_generated"

func Test_no_changes(t *testing.T) {
	// Given
	dir := initTestGit("no_changes")
	// When
	changes := services.ChangedFiles(dir)
	// Then
	i := is.New(t)
	i.True(len(changes) == 0)
}

func Test_one_new_file(t *testing.T) {
	// Given
	dir := initTestGit("one_new_file")
	touchFile(dir, "logo.svg")
	// When
	changes := services.ChangedFiles(dir)
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal("logo.svg", changes[0].Path)
}

func Test_multiple_new_files(t *testing.T) {
	// Given
	dir := initTestGit("multiple_new_files")
	touchFile(dir, "logo.png")
	touchFile(dir, "logo.svg")
	// When
	changes := services.ChangedFiles(dir)
	// Then
	i := is.New(t)
	i.True(len(changes) == 2)
	i.Equal("logo.png", changes[0].Path)
	i.Equal("logo.svg", changes[1].Path)
}

func Test_new_path_with_file(t *testing.T) {
	// Given
	dir := initTestGit("new_path_with_file")
	touchFile(dir, "images/logo.svg")
	// When
	changes := services.ChangedFiles(dir)
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal("images/logo.svg", changes[0].Path)
}

func Test_file_with_capital_letter(t *testing.T) {
	// Given
	dir := initTestGit("file_with_capital_letter")
	touchFile(dir, "images/Logo.svg")
	// When
	changes := services.ChangedFiles(dir)
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal("images/Logo.svg", changes[0].Path)
}

func Test_status_untracked(t *testing.T) {
	// Given
	dir := initTestGit("status_untracked")
	touchFile(dir, "Logo.svg")
	// When
	changes := services.ChangedFiles(dir)
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal(services.StatusUntracked, changes[0].UnstagedStatus)
}

func Test_status_added(t *testing.T) {
	// Given
	dir := initTestGit("status_added")
	touchFile(dir, "logo.svg")
	gitAdd(dir, "logo.svg")
	// When
	changes := services.ChangedFiles(dir)
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal(services.StatusAdded, changes[0].UnstagedStatus)
}

// @todo test all characters in filename

func initTestGit(testDir string) string {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	dir := currentDir + "/" + mockDir + "/" + testDir
	// Clean up directory from old test
	_, err = os.Stat(dir)
	if !os.IsNotExist(err) {
		err := os.RemoveAll(dir)
		if err != nil {
			log.Fatal(err)
		}
	}
	// Create directory
	err = os.Mkdir(dir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	// Create empty git repository
	cmd := exec.Command("git", "init", dir)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	err = os.Chdir(dir)
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func touchFile(dir string, fileName string) {
	// Ensure dir exists
	subDir := path.Dir(fileName)
	fullDir := filepath.Join(dir, subDir)
	os.MkdirAll(fullDir, 0755)
	// Create file
	file, err := os.Create(filepath.Join(dir, fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
}

func gitAdd(dir string, file string) {
	// Add file or directory to the stage
	_, err := services.GitAdd(filepath.Join(dir, file))
	if err != nil {
		log.Fatal(err)
	}
}
