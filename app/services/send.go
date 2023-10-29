package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/confetti-framework/errors"
	"github.com/confetti-framework/framework/inter"
	"github.com/spf13/cast"
)

var UserError = errors.New("something went wrong, you can probably adjust it yourself to fix it")

var retry = 0

func Send(cli inter.Cli, url string, body any, method string, env Environment, repo string) (string, error) {
	token, err := GetAccessToken(cli, env)
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
	req.Header.Add("Authorization", "Bearer "+token)
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
		return "", fmt.Errorf("error reading response body: %w", err)
	}
	if res.StatusCode == http.StatusForbidden {
		if retry < 4 {
			fmt.Printf("\rSetting up development services. This usually takes 5 seconds")
		} else {
			fmt.Printf("\rOperation is taking longer than expected.                    ")
		}
		if retry == 0 {
			errStartDevContainers := startDevContainers(env, repo)
			if errStartDevContainers != nil {
				return "", fmt.Errorf("error starting dev containers: %w", errStartDevContainers)
			}
		}
		time.Sleep(1 * time.Second)
		retry++
		return Send(cli, url, body, method, env, repo)
	}
	if res.StatusCode == http.StatusBadGateway {
		// Override previous message with spaces
		fmt.Printf("\rService is almost available. We'll be done in 2 seconds       ")
		time.Sleep(1 * time.Second)
		retry++
		return Send(cli, url, body, method, env, repo)
	}
	retry = 0
	if res.StatusCode > 299 {
		if res.StatusCode == http.StatusBadRequest {
			body := map[string]interface{}{}
			err := json.Unmarshal(responseBody, &body)
			if err != nil {
				return "", fmt.Errorf("error unmarshalling response body: %w", err)
			}
			errs := body["errors"].([]any)
			title := errs[0].(map[string]any)["title"].(string)
			return string(responseBody), fmt.Errorf("%w: %s", UserError, title)
		}
		err := errors.New(
			"error with status: " + cast.ToString(res.StatusCode) +
				" with request: " + method + " " + url +
				" and response: " + string(responseBody),
		)
		return string(responseBody), err
	}
	return string(responseBody), nil
}

func startDevContainers(env Environment, repository string) error {
	jsonData := map[string]string{
		"environment_key": env.Key,
		"repository":      repository,
	}
	jsonValue, _ := json.Marshal(jsonData)
	response, err := http.Post(
		env.GetOrchestratorApi() + "/start_development",
		"application/json",
		bytes.NewBuffer(jsonValue),
	)
	defer response.Body.Close()

	if err != nil {
		bodyString := ""
		if response.Body != nil {
			bodyBytes, _ := ioutil.ReadAll(response.Body)
			bodyString = string(bodyBytes)
		}
		return fmt.Errorf("error: %v, response: %s", err, bodyString)
	}
	return nil
}
