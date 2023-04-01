package services

import (
	"net/http"
    "path/filepath"
    "strings"
)

func SendDeleteSource(path string) error {
    // Ignore hidden files and directories
    if strings.HasPrefix(path, ".") || strings.HasPrefix(filepath.Base(path), ".") {
        return nil
    }
	_, err := Send("http://api.localhost/parser/source?path="+path, "", http.MethodDelete)
    return err
}
