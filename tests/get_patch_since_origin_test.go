package test

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
	patch, err := services.GetPatchSinceOrigin(path.Join(dir, "unkown.md"))
	// Then
	i := is.New(t)
	i.Equal(patch, "")
	i.Equal(true, strings.Contains(err.Error(), "unknown revision or path not in the working tree"))
	i.Equal(false, strings.Contains(err.Error(), "\n"))
}

func Test_patch_without_head(t *testing.T) {
	// Given
	dir := initTestGit()
	touchFile(dir, "readme.md")
	setFileContent(dir, "readme.md", "<?php")
	// When
	patch, _ := services.GetPatchSinceOrigin(path.Join(dir, "readme.md"))
	// Then
	i := is.New(t)
	i.Equal("", patch)
	// todo: get error "git repository not exists or nog connected: unknown HEAD."
	// i.Equal(nil, err)
}

func Test_patch_one_line_added(t *testing.T) {
	// Given
	dir := initTestGit()
	touchFile(dir, "readme.md")
	setFileContent(dir, "readme.md", "<?php")
	// When
	patch, err := services.GetPatchSinceOrigin(path.Join(dir, "readme.md"))
	// Then
	i := is.New(t)
	i.NoErr(err)
	i.Equal("+ <?php", patch)
}
