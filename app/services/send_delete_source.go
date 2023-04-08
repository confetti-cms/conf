package services

import (
	"net/http"
	"path/filepath"
	"src/config"
	"strings"
)

func SendDeleteSource(path string) error {
	// Ignore hidden files and directories
	if strings.HasPrefix(path, ".") || strings.HasPrefix(filepath.Base(path), ".") {
		return nil
	}
	host := config.App.Host
	_, err := Send("http://api." + host + "/parser/source?path="+path, "", http.MethodDelete)
	return err
}
