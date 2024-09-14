package services

import (
	"net/http"

	"github.com/confetti-framework/framework/inter"
)

type ParseComponentBody struct {
	File string `json:"file"`
}

func ParseComponent(cli inter.Cli, env Environment, body ParseComponentBody, repo string) error {
	url := env.GetServiceUrl("confetti-cms/parser")
	_, err := Send(cli, url+"/parse_component", body, http.MethodPost, env, repo)
	return err
}
