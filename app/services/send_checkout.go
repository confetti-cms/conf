package services

import (
	"net/http"
)

type CheckoutBody struct {
	Commit string `json:"commit"`
	Reset  bool   `json:"reset"`
}

func SendCheckout(requestBody CheckoutBody) error {
	return Send("http://api.localhost/parser/checkout", requestBody, http.MethodPut)
}
