package services

import (
	"bytes"
	"encoding/json"
	"github.com/confetti-framework/errors"
	"github.com/spf13/cast"
	"io/ioutil"
	"net/http"
	"time"
)

func Send(url string, body any, method string) (string, error) {
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
	req.Header.Add("Accept-Language", "application/json")
	req.Header.Add("Content-Type", "application/json")
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
				" with request url: " + url +
				" and response: " + string(responseBody),
		)
	}
	return string(responseBody), nil
}
