package services

import (
	"net/http"
	"time"

	"github.com/confetti-framework/framework/inter"
)

func ParseBaseComponents(cli inter.Cli, env Environment, repo string) error {
	url := env.GetServiceUrl("confetti-cms/parser")
	_, err := Send(cli, url+"/parse_base_components", "", http.MethodPost, env, repo, 30*time.Second)
	return err
}
