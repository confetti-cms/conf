package services

import (
	"net/http"
    "path/filepath"
    "strings"
)

func SendDeleteSource(path string) error {
    println("Delete source: " + path)
    // Ignore hidden files and directories
    if strings.HasPrefix(path, ".") || strings.HasPrefix(filepath.Base(path), ".") {
        return nil
    }
	return Send("http://api.localhost/parser/source?path="+path, "", http.MethodDelete)
}
