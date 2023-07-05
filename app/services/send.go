package services

import (
	"bytes"
	"encoding/json"
	"github.com/confetti-framework/errors"
	"github.com/confetti-framework/framework/inter"
	"github.com/spf13/cast"
	"io/ioutil"
	"net/http"
	"src/app/services/auth"
	"time"
)

func Send(cli inter.Cli, url string, body any, method string) (string, error) {
	token, err := auth.GetAccessToken(cli)
	if err != nil {
		return "", err
	}
	payloadB, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	payload := bytes.NewBuffer(payloadB)
	// Create request
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return "", err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer " + token)
	// Do request
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	// Create response
	responseBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		println("error response: " + string(responseBody))
		return "", err
	}
	if res.StatusCode > 299 {
		return string(responseBody), errors.New(
			"error with status: " + cast.ToString(res.StatusCode) +
				" with request: " + method + " " + url +
				" and response: " + string(responseBody),
		)
	}
	return string(responseBody), nil
}
