package services

import (
	"net/http"
)

type CheckoutBody struct {
	Commit string `json:"commit"`
}

func SendCheckout(requestBody CheckoutBody) error {
	return Send("http://api.localhost/parser/checkout", requestBody, http.MethodPut)
}
