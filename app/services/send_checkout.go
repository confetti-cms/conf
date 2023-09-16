package services

import (
	"github.com/confetti-framework/framework/inter"
	"net/http"
)

type CheckoutBody struct {
	Commit string `json:"commit"`
	Reset  bool   `json:"reset"`
}

func SendCheckout(cli inter.Cli, env Environment, requestBody CheckoutBody, repo string) error {
	url := env.GetServiceUrl("confetti-cms/parser")
	_, err := Send(cli, url + "/checkout", requestBody, http.MethodPut, env, repo)
    return err
}
