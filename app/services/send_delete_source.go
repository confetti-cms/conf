package services

import (
	"net/http"
	"time"

	"github.com/confetti-framework/framework/inter"
)

func SendDeleteSource(cli inter.Cli, env Environment, path string, repo string) error {
	url := env.GetServiceUrl("confetti-cms/parser")
	_, err := Send(cli, url+"/source?path="+path, "", http.MethodDelete, env, repo, 30*time.Second)
	return err
}
