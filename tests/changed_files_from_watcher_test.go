package tests

import (
	"src/app/services"
	"testing"

	"github.com/matryer/is"
)

func Test_no_changes(t *testing.T) {
	// Given
	dir := initTestGit()

	// When
	changes := services.ChangedFilesSinceRemoteCommit(dir, "")
	// Then
	i := is.New(t)
	i.True(len(changes) == 0)
}

func Test_one_new_file(t *testing.T) {
	// Given
	dir := initTestGit()
	touchFile(dir, "logo.svg")
	// When
	changes := services.ChangedFilesSinceRemoteCommit(dir, "")
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal("logo.svg", changes[0].Path)
}

func Test_multiple_new_files(t *testing.T) {
	// Given
	dir := initTestGit()
	touchFile(dir, "logo.png")
	touchFile(dir, "logo.svg")
	// When
	changes := services.ChangedFilesSinceRemoteCommit(dir, "")
	// Then
	i := is.New(t)
	i.True(len(changes) == 2)
	i.Equal("logo.png", changes[0].Path)
	i.Equal("logo.svg", changes[1].Path)
}

func Test_new_path_with_file(t *testing.T) {
	// Given
	dir := initTestGit()
	touchFile(dir, "images/logo.svg")
	// When
	changes := services.ChangedFilesSinceRemoteCommit(dir, "")
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal("images/logo.svg", changes[0].Path)
}

func Test_file_with_capital_letter(t *testing.T) {
	// Given
	dir := initTestGit()
	touchFile(dir, "images/Logo.svg")
	// When
	changes := services.ChangedFilesSinceRemoteCommit(dir, "")
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal("images/Logo.svg", changes[0].Path)
}

func Test_file_with_number(t *testing.T) {
	// Given
	dir := initTestGit()
	touchFile(dir, "images/Logo2.svg")
	// When
	changes := services.ChangedFilesSinceRemoteCommit(dir, "")
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal("images/Logo2.svg", changes[0].Path)
}

func Test_file_with_special_caracters(t *testing.T) {
	// Given
	dir := initTestGit()
	touchFile(dir, "images/Logo-_.svg")
	// When
	changes := services.ChangedFilesSinceRemoteCommit(dir, "")
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal("images/Logo-_.svg", changes[0].Path)
}

func Test_status_staged_added(t *testing.T) {
	// Given
	dir := initTestGit()
	touchFile(dir, "logo.svg")
	gitAdd(dir, "logo.svg")
	// When
	changes := services.ChangedFilesSinceRemoteCommit(dir, "")
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal(services.GitStatusAdded, changes[0].Status)
}

func Test_status_staged_modified(t *testing.T) {
	// Given
	dir := initTestGit()
	touchFile(dir, "logo.svg")
	gitAdd(dir, "logo.svg")
	gitCommit(dir, "logo.svg")
	setFileContent(dir, "logo.svg", "Content")
	gitAdd(dir, "logo.svg")
	// When
	changes := services.ChangedFilesSinceRemoteCommit(dir, "")
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal(services.GitStatusModified, changes[0].Status)
}

func Test_status_staged_deleted(t *testing.T) {
	// Given
	dir := initTestGit()
	touchFile(dir, "logo.svg")
	setFileContent(dir, "logo.svg", "Content")
	gitAdd(dir, "logo.svg")
	gitCommit(dir, "logo.svg")
	deleteFile(dir, "logo.svg")
	gitAdd(dir, "logo.svg")
	// When
	changes := services.ChangedFilesSinceRemoteCommit(dir, "")
	// Then
	i := is.New(t)
	i.True(len(changes) == 1)
	i.Equal(services.GitStatusDeleted, changes[0].Status)
}

func Test_status_staged_renamed(t *testing.T) {
	// Given
	dir := initTestGit()
	touchFile(dir, "logo1.svg")
	setFileContent(dir, "logo1.svg", `The content`)
	gitAdd(dir, "logo1.svg")
	gitCommit(dir, "logo1.svg")
	deleteFile(dir, "logo1.svg")
	touchFile(dir, "logo2.svg")
	setFileContent(dir, "logo2.svg", `The content`)
	gitAdd(dir, "logo1.svg") // Also add deleted file
	gitAdd(dir, "logo2.svg")
	// When
	changes := services.ChangedFilesSinceRemoteCommit(dir, "")
	// Then
	i := is.New(t)
	i.True(len(changes) == 2)
	i.Equal("logo1.svg", changes[0].Path)
	i.Equal(services.GitStatusDeleted, changes[0].Status)
	i.Equal("logo2.svg", changes[1].Path)
	i.Equal(services.GitStatusRenamed, changes[1].Status)
}
