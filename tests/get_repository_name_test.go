package tests

import (
	"src/app/services"
	"testing"

	"github.com/matryer/is"
)

func Test_get_repository_by_empty_url(t *testing.T) {
	// Given
	url := ""

	// When
	name, err := services.GetRepositoryNameByOriginUrl(url)

	// Then
	i := is.New(t)
	i.Equal(name, "")
	i.True(err != nil)
}

func Test_get_repository_by_ssh_url(t *testing.T) {
	// Given
	url := "git@github.com:confetti-cms/office.git"

	// When
	name, err := services.GetRepositoryNameByOriginUrl(url)

	// Then
	i := is.New(t)
	i.Equal(name, "confetti-cms/office")
	i.NoErr(err)
}

func Test_get_repository_by_https_url(t *testing.T) {
	// Given
	url := "https://github.com/confetti-cms/office.git"

	// When
	name, err := services.GetRepositoryNameByOriginUrl(url)

	// Then
	i := is.New(t)
	i.Equal(name, "confetti-cms/office")
	i.NoErr(err)
}
