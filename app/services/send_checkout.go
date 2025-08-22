package services

import (
	"net/http"
	"time"

	"github.com/confetti-framework/framework/inter"
)

type CheckoutBody struct {
	Commit string `json:"commit"`
	Reset  bool   `json:"reset"`
	Parse  bool   `json:"parse"`
}

func SendCheckout(cli inter.Cli, env Environment, requestBody CheckoutBody, repo string) error {
	url := env.GetServiceUrl("confetti-cms/parser")
	_, err := Send(cli, url+"/checkout", requestBody, http.MethodPut, env, repo, 30*time.Second)
	return err
}
