package services

import (
	"github.com/confetti-framework/framework/inter"
	"net/http"
	"src/config"
)

type CheckoutBody struct {
	Commit string `json:"commit"`
	Reset  bool   `json:"reset"`
}

func SendCheckout(cli inter.Cli, requestBody CheckoutBody) error {
	host := config.App.Host
	_, err := Send(cli, "http://api." + host + "/parser/checkout", requestBody, http.MethodPut)
    return err
}
