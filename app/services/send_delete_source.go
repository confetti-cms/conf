package services

import (
	"github.com/confetti-framework/framework/inter"
	"net/http"
	"src/config"
)

func SendDeleteSource(cli inter.Cli, path string) error {
	host := config.App.Host
	_, err := Send(cli, "http://api." + host + "/parser/source?path="+path, "", http.MethodDelete)
	return err
}
