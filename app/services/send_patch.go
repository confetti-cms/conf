package services

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type PatchBody struct {
	Path      string `json:"path"`
	Patch     string `json:"patch"`
	Untracked bool   `json:"is_untracked"`
}

func SendPatch(requestBody PatchBody) error {
	url := "http://api.localhost/parser/source"
	method := "PATCH"
	payloadB, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	payload := bytes.NewReader(payloadB)
	// Create request
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return err
	}
	req.Header.Add("Accept-Language", "application/json")
	req.Header.Add("Content-Type", "application/json")
	// Do request
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	// Create response
	responseBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		println("Error response: " + string(responseBody))
		return err
	}
	return nil
}
