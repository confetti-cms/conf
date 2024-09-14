package services

import (
	"net/http"

	"github.com/confetti-framework/framework/inter"
)

func ParseBaseComponents(cli inter.Cli, env Environment, repo string) error {
	url := env.GetServiceUrl("confetti-cms/parser")
	_, err := Send(cli, url+"/parse_base_components", "", http.MethodPost, env, repo)
	return err
}
