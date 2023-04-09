package services

import (
	"net/http"
	"src/config"
)

func SendDeleteSource(path string) error {
	host := config.App.Host
	_, err := Send("http://api." + host + "/parser/source?path="+path, "", http.MethodDelete)
	return err
}
