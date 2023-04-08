package services

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path"
	"src/config"
)

const hiddenDir = ".confetti"

func GetHiddenFiles(root string, verbose bool) error {
	// Get content of component
	host := config.App.Host
	body, err := Send("http://api." + host + "/parser/source/components", nil, http.MethodGet)
	if err != nil {
		return err
	}
	// Get file content from response
	contentsRaw := []map[string]string{}
	json.Unmarshal([]byte(body), &contentsRaw)
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
		if verbose {
			println("Standard hidden component saved: " + target)
		}
	}
	return nil
}
