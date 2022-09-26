package tests

import (
	"path"
	"src/app/services"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func Test_patch_unkown_file(t *testing.T) {
	// Given
	dir := initTestGit()
	// When
	patch, err := services.GetPatchSinceCommit("", path.Join(dir, "unkown.md"))
	// Then
	i := is.New(t)
	i.Equal(patch, "")
	i.NoErr(err)
}

func Test_patch_line_changed(t *testing.T) {
	// Given
	dir := initTestGit()
	file := "readme.md"
	touchFile(dir, file)
	gitAdd(dir, file)
	gitCommit(dir, file)
	setFileContent(dir, file, "<?php")
	// When
	patch, err := services.GetPatchSinceCommit("", path.Join(dir, file))
	// Then
	i := is.New(t)
	i.NoErr(err)
	i.True(strings.Contains(patch, "--- a/readme.md"))
	i.True(strings.Contains(patch, "+++ b/readme.md"))
	i.True(strings.Contains(patch, "No newline at end of file"))
	i.True(strings.Contains(patch, "@@ -0,0 +1 @@\n+<?php\n\\ No newline at end of file"))
}

func Test_patch_from_specific_commit(t *testing.T) {
	// Given
	dir := initTestGit()
	file := "readme.md"
	touchFile(dir, file)
	gitAdd(dir, file)
	gitCommit(dir, file)
	setFileContent(dir, file, "first_")
	gitAdd(dir, file)
	gitCommit(dir, file)
	setFileContent(dir, file, "second_")
	since := getCommitFromLog(dir, 1)
	// When
	patch, err := services.GetPatchSinceCommit(since, path.Join(dir, file))
	// Then
	i := is.New(t)
	i.NoErr(err)
	i.True(strings.Contains(patch, "+first_second_")) // Contains
	i.True(!strings.Contains(patch, "-first_"))       // Not contains
}
