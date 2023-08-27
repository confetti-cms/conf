package services

import (
	"net/http"

	"github.com/confetti-framework/framework/inter"
)

func SendDeleteSource(cli inter.Cli, env Environment, path string) error {
	url := env.GetServiceUrl("confetti-cms/parser")
	_, err := Send(cli, url+"/source?path="+path, "", http.MethodDelete)
	return err
}
