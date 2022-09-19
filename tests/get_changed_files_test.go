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

func Test_file_with_number(t *testing.T) {
	// Given
	dir := initTestGit("file_with_capital_letter")
	touchFile(dir, "images/Logo2.svg")
	// When
	changes := services.ChangedFiles(dir)
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal("images/Logo2.svg", changes[0].Path)
}

func Test_file_with_special_caracters(t *testing.T) {
	// Given
	dir := initTestGit("file_with_capital_letter")
	touchFile(dir, "images/Logo-_.svg")
	// When
	changes := services.ChangedFiles(dir)
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal("images/Logo-_.svg", changes[0].Path)
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

func Test_status_unstaged_modified(t *testing.T) {
	// Given
	dir := initTestGit("unstaged_modified")
	touchFile(dir, "logo.svg")
	gitAdd(dir, "logo.svg")
	gitCommit(dir, "logo.svg")
	setFileContent(dir, "logo.svg", "Content")
	// When
	changes := services.ChangedFiles(dir)
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal(services.StatusModified, changes[0].UnstagedStatus)
}

func Test_status_unstaged_deleted(t *testing.T) {
	// Given
	dir := initTestGit("unstaged_deleted")
	touchFile(dir, "logo.svg")
	setFileContent(dir, "logo.svg", "Content")
	gitAdd(dir, "logo.svg")
	gitCommit(dir, "logo.svg")
	deleteFile(dir, "logo.svg")
	// When
	changes := services.ChangedFiles(dir)
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal(services.StatusDeleted, changes[0].UnstagedStatus)
}

func Test_status_staged_added(t *testing.T) {
	// Given
	dir := initTestGit("staged_added")
	touchFile(dir, "logo.svg")
	gitAdd(dir, "logo.svg")
	// When
	changes := services.ChangedFiles(dir)
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal(services.StatusUnchanged, changes[0].UnstagedStatus)
	i.Equal(services.StatusAdded, changes[0].StagedStatus)
}

func Test_status_staged_modified(t *testing.T) {
	// Given
	dir := initTestGit("staged_modified")
	touchFile(dir, "logo.svg")
	gitAdd(dir, "logo.svg")
	gitCommit(dir, "logo.svg")
	setFileContent(dir, "logo.svg", "Content")
	gitAdd(dir, "logo.svg")
	// When
	changes := services.ChangedFiles(dir)
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal(services.StatusUnchanged, changes[0].UnstagedStatus)
	i.Equal(services.StatusModified, changes[0].StagedStatus)
}

func Test_status_staged_deleted(t *testing.T) {
	// Given
	dir := initTestGit("staged_deleted")
	touchFile(dir, "logo.svg")
	setFileContent(dir, "logo.svg", "Content")
	gitAdd(dir, "logo.svg")
	gitCommit(dir, "logo.svg")
	deleteFile(dir, "logo.svg")
	gitAdd(dir, "logo.svg")
	// When
	changes := services.ChangedFiles(dir)
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal(services.StatusUnchanged, changes[0].UnstagedStatus)
	i.Equal(services.StatusDeleted, changes[0].StagedStatus)
}

func Test_status_staged_renamed(t *testing.T) {
	// Given
	dir := initTestGit("status_staged_renamed")
	touchFile(dir, "logo1.svg")
	setFileContent(dir, "logo1.svg", `As you noticed, Git performs rename detection using a heuristic, rather than being told that a rename occurred. The git mv command, in fact, simply stages an add on the new file path and a remove of the old file path. Thus, rename detection is performed by comparing the contents of added files to the previously committed contents of deleted files.\nFirst, candidates are collected. Any new files are possible rename targets and any deleted files are possible rename sources. In addition, rewriting changes are broken such that a file that is more than 50 different than its previous revision is both a possible rename source and a possible rename target.\nNext, identical renames are detected. If you rename a file without making any changes, then the file will hash identically. These can be detected just performing comparisons of the hash in the index without reading the file contents, so removing these from the candidate list will reduce the number of comparisons you need to perform.`)
	gitAdd(dir, "logo1.svg")
	gitCommit(dir, "logo1.svg")
	deleteFile(dir, "logo1.svg")
	touchFile(dir, "logo2.svg")
	setFileContent(dir, "logo2.svg", `As you noticed, Git performs rename detection using a heuristic, rather than being told that a rename occurred. The git mv command, in fact, simply stages an add on the new file path and a remove of the old file path. Thus, rename detection is performed by comparing the contents of added files to the previously committed contents of deleted files.\nFirst, candidates are collected. Any new files are possible rename targets and any deleted files are possible rename sources. In addition, rewriting changes are broken such that a file that is more than 50 different than its previous revision is both a possible rename source and a possible rename target.\nNext, identical renames are detected. If you rename a file without making any changes, then the file will hash identically. These can be detected just performing comparisons of the hash in the index without reading the file contents, so removing these from the candidate list will reduce the number of comparisons you need to perform.`)
	gitAdd(dir, "logo1.svg") // Also add deleted file
	gitAdd(dir, "logo2.svg")
	// When
	changes := services.ChangedFiles(dir)
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal("logo2.svg", changes[0].Path)
	i.Equal("logo1.svg", changes[0].FromPath)
	i.Equal(services.StatusUnchanged, changes[0].UnstagedStatus)
	i.Equal(services.StatusRenamed, changes[0].StagedStatus)
}

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
		log.Fatalf("failed to run `git init`: %s", err)
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
		log.Fatalf("failed to run `git add`: %s", err)
	}
}

func gitCommit(dir string, file string) {
	// Commit file
	_, err := services.GitCommit(filepath.Join(dir, file))
	if err != nil {
		log.Fatalf("failed to run `git commit`: %s", err)
	}
}

func setFileContent(dir string, fileName string, content string) {
	file, err := os.OpenFile(filepath.Join(dir, fileName), os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		log.Fatalf("failed writing to file: %s", err)
	}
}

func deleteFile(dir string, fileName string) {
	e := os.Remove(filepath.Join(dir, fileName))
	if e != nil {
		log.Fatal(e)
	}
}
