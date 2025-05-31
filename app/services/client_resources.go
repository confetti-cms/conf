package services

import (
	"encoding/json"
	"fmt"
	"net/url"
	"src/config"
	"time"

	"net/http"
	"os"
	"path"
	"strings"

	"github.com/confetti-framework/framework/inter"
)

const sharedResourcesDir = ".confetti"

const ComponentClassSuffix = "Component.php"

func IsBaseComponent(file string) bool {
	return strings.HasSuffix(file, ComponentClassSuffix)
}

func RemoveAllLocalResources() error {
	err := os.RemoveAll(path.Join(config.Path.Root, sharedResourcesDir))
	if err != nil {
		return fmt.Errorf("failed to remove all local resources: %w", err)
	}
	return nil
}

func FetchResources(cli inter.Cli, env Environment, repo string, since time.Time) error {
	// Get content of component
	files, err := getResourceFileNames(cli, env, repo, since)
	if err != nil {
		return fmt.Errorf("can't fetch file names: %w", err)
	}
	// Remove files with '.removed' suffix
	for _, file := range files {
		if strings.HasSuffix(file, ".removed") {
			if config.App.VeryVerbose {
				println("Remove resource file: " + file)
			}
			err := removeResourceFile(file)
			if err != nil {
				fmt.Println("can't fetch resource file: %w", err)
			}
			if config.App.VeryVerbose {
				fmt.Printf("file removed: %s\n", file)
			}
		}
	}
	// Fetch and save the remaining files
	for _, file := range files {
		if !strings.HasSuffix(file, ".removed") {
			err := fetchAndSaveResourceFiles(cli, env, repo, file)
			if err != nil {
				return fmt.Errorf("failed to fetch and save resource files: %w", err)
			}
			if config.App.VeryVerbose {
				fmt.Printf("file saved: %s\n", file)
			}
		}
	}
	return nil
}

func getResourceFileNames(cli inter.Cli, env Environment, repo string, since time.Time) ([]string, error) {
	baseUrl := env.GetServiceUrl("confetti-cms/shared-resource")
	content, err := Send(cli, baseUrl+"/resources?"+sinceParameter(since), nil, http.MethodGet, env, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resource file names: %w", err)
	}
	var files []string
	json.Unmarshal([]byte(content), &files)
	if err != nil {
		return nil, fmt.Errorf("unable to decode JSON response: %w", err)
	}
	return files, nil
}

func removeResourceFile(target string) error {
	target = strings.TrimSuffix(target, ".removed")
	target = path.Join(config.Path.Root, sharedResourcesDir, target)
	err := os.Remove(target)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove file: %w", err)
	}
	if config.App.VeryVerbose {
		println("File removed: " + target)
	}
	return nil
}

func fetchAndSaveResourceFiles(cli inter.Cli, env Environment, repo, file string) error {
	baseUrl := env.GetServiceUrl("confetti-cms/shared-resource")
	content, err := Send(cli, baseUrl+"/resources/content?file="+url.QueryEscape(file), nil, http.MethodGet, env, repo)
	if err != nil {
		return fmt.Errorf("failed to fetch content of resource: %w", err)
	}
	// Save hidden component
	target := path.Join(config.Path.Root, sharedResourcesDir, file)
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
	if config.App.VeryVeryVerbose {
		println("Resource fetched and saved: " + target)
	}
	return nil
}

func sinceParameter(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return "date_since=" + url.QueryEscape(t.UTC().Format("2006-01-02 15:04:05"))
}
