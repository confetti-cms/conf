package services

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path"
	"src/config"
	"strings"

	"github.com/confetti-framework/framework/inter"
)

const hiddenDir = ".confetti"

// ComponentConfigSuffix
// Actually, there should be another letter 'c' as the first letter here,
// but we don't consider it because it can be in lowercase or uppercase.
const ComponentConfigSuffix = "omponent.blade.php"
const ComponentClassSuffix = "omponent.class.php"

func IsHiddenFileGenerator(file string) bool {
	return strings.HasSuffix(file, ComponentConfigSuffix) || strings.HasSuffix(file, ComponentClassSuffix)
}

func FetchHiddenFiles(cli inter.Cli, env Environment, root string, repo string) error {
	// Get content of component
	url := env.GetServiceUrl("confetti-cms/parser")
	body, err := Send(cli, url+"/source/components", nil, http.MethodGet, env, repo)
	if err != nil {
		return err
	}
	err = os.RemoveAll(path.Join(root, hiddenDir))
	if err != nil {
		return err
	}
	// Get file content from response
	contentsRaw := []map[string]string{}
	err = json.Unmarshal([]byte(body), &contentsRaw)
	if err != nil {
		return err
	}
	for _, contentRaw := range contentsRaw {
		content64 := contentRaw["content"]
		file := contentRaw["file"]
		content, err := base64.StdEncoding.DecodeString(content64)
		if err != nil {
			return err
		}
		// Save hidden component
		target := path.Join(root, hiddenDir, file)
		err = os.MkdirAll(path.Dir(target), os.ModePerm)
		if err != nil {
			return err
		}
		f, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(string(content))
		if err != nil {
			return err
		}
		if config.App.Debug {
			println("Standard hidden component saved: " + target)
		}
	}
	return nil
}
