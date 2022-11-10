package services

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type PatchBody struct {
	Path      string `json:"path"`
	Patch     string `json:"patch"`
	Untracked bool   `json:"is_untracked"`
}

func SendPatch(requestBody PatchBody) error {
	url := "http://localhost:3001/api/source"
	method := "PATCH"
	json, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	payload := bytes.NewReader(json)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return err
	}
	req.Header.Add("Accept-Language", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	responseBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		println("Error response: " + string(responseBody))
		return err
	}

	return nil
}
