package services

import (
	"net/http"
	"time"

	"github.com/confetti-framework/framework/inter"
)

type ParseComponentBody struct {
	File string `json:"file"`
}

func ParseComponent(cli inter.Cli, env Environment, body ParseComponentBody, repo string) error {
	url := env.GetServiceUrl("confetti-cms/parser")
	_, err := Send(cli, url+"/parse_component", body, http.MethodPost, env, repo, 30*time.Second)
	return err
}

func ParseAllComponents(cli inter.Cli, env Environment, repo string) error {
	url := env.GetServiceUrl("confetti-cms/parser")
	_, err := Send(cli, url+"/parse_all_components", []string{}, http.MethodPost, env, repo, 30*time.Second)
	return err
}
