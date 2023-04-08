package services

import (
	"net/http"
	"src/config"
)

type CheckoutBody struct {
	Commit string `json:"commit"`
	Reset  bool   `json:"reset"`
}

func SendCheckout(requestBody CheckoutBody) error {
	host := config.App.Host
	_, err := Send("http://api." + host + "/parser/checkout", requestBody, http.MethodPut)
    return err
}
