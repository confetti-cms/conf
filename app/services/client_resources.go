package services

import (
	"encoding/json"
	"fmt"
	"net/url"
	"src/config"
	"time"

	//	"encoding/base64"
	//	"encoding/json"
	"net/http"
	"os"
	"path"
	//	"src/config"
	"strings"

	"github.com/confetti-framework/framework/inter"
)

const hiddenDir = ".confetti"

// ComponentConfigSuffix
// Actually, there should be another letter 'c' as the first letter here,
// but we don't consider it because it can be in lowercase or uppercase.
const ComponentConfigSuffix = "omponent.blade.php"
const ComponentClassSuffix = "omponent.class.php"

func IsComponentFileGenerator(file string) bool {
	return strings.HasSuffix(file, ComponentConfigSuffix) || strings.HasSuffix(file, ComponentClassSuffix)
}

// GenerateComponentFiles only generate the files on the host, on an other place we fetch the resource files
func GenerateComponentFiles(cli inter.Cli, env Environment, repo string) error {
	// Get content of component
	url := env.GetServiceUrl("confetti-cms/parser")
	_, err := Send(cli, url+"/source/components", nil, http.MethodGet, env, repo)
	if err != nil {
		return fmt.Errorf("failed to generate component files on host: %w", err)
	}
	return nil
}

func RemoveAllClientResources() error {
	err := os.RemoveAll(path.Join(config.Path.Root, hiddenDir))
	if err != nil {
		return fmt.Errorf("failed to remove all client resources: %w", err)
	}
	return nil
}

type FetchResourcesSince struct {
	AllTime bool
	Time    time.Time
}

func (f FetchResourcesSince) parameter() string {
	if f.AllTime {
		return ""
	}
	return "date_since=" + url.QueryEscape(f.Time.UTC().Format("2006-01-02 15:04"))
}

func FetchResources(cli inter.Cli, env Environment, repo string, since FetchResourcesSince) error {
	// Get content of component
	files, err := getResourceFileNames(cli, env, repo, since)
	if err != nil {
		return err
	}
	// Fetch and save
	for _, file := range files {
		err2 := fetchAndSaveResourceFiles(cli, env, repo, file)
		if err2 != nil {
			return fmt.Errorf("failed to fetch and save resource files: %w", err2)
		}
	}
	return nil
}

func getResourceFileNames(cli inter.Cli, env Environment, repo string, since FetchResourcesSince) ([]string, error) {
	baseUrl := env.GetServiceUrl("confetti-cms/shared-resource")
	content, err := Send(cli, baseUrl+"/resources?"+since.parameter(), nil, http.MethodGet, env, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resource files: %w", err)
	}
	var files []string
	json.Unmarshal([]byte(content), &files)
	if err != nil {
		return nil, fmt.Errorf("unable to decode JSON response: %w", err)
	}
	return files, nil
}

func fetchAndSaveResourceFiles(cli inter.Cli, env Environment, repo, file string) error {
	baseUrl := env.GetServiceUrl("confetti-cms/shared-resource")
	content, err := Send(cli, baseUrl+"/resources/content?file="+url.QueryEscape(file), nil, http.MethodGet, env, repo)
	if err != nil {
		return fmt.Errorf("failed to fetch resource files: %w", err)
	}
	// Save hidden component
	target := path.Join(config.Path.Root, hiddenDir, file)
	err = os.MkdirAll(path.Dir(target), os.ModePerm)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {
		return err
	}
	if config.App.Debug {
		println("Resource fetched and saved: " + target)
	}
	return nil
}
