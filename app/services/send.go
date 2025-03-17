package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"src/config"
	"time"

	"io"

	"github.com/confetti-framework/errors"
	"github.com/confetti-framework/framework/inter"
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
	debugRequest(method, url, payload.String())
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
	responseBody, err := io.ReadAll(io.Reader(res.Body))
	if err != nil {
		println("error response: " + string(responseBody))
		return "", fmt.Errorf("error reading response body: %w", err)
	}
	// Use retry mechanism to wait until development containers are up and running
	// Retries are set up to 4 times, each with a delay of 1 second
	// If the operation is longer than expected, an informative message is displayed to the user
	if res.StatusCode == http.StatusForbidden {
		if retry < 6 {
			fmt.Printf("\rSetting up development services. This usually takes 5 seconds    ")
		} else {
			fmt.Printf("\rOperation is taking longer than expected.                        ")
		}

		errStartDevContainers := startDevContainers(env, repo)
		if errStartDevContainers != nil {
			if retry > 20 {
				return "", fmt.Errorf("error starting dev containers: %w", errStartDevContainers)
			}
		}
		time.Sleep(1 * time.Second)
		retry++
		return Send(cli, url, body, method, env, repo)
	}
	if res.StatusCode == http.StatusBadGateway {
		// Override previous message with spaces
		fmt.Printf("\rDevelopment services are almost available. We'll be done in 3 seconds")
		if config.App.VeryVerbose {
			fmt.Println("Body:", string(responseBody))
		}
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
		err := fmt.Errorf(
			"error with status: %d with request: %s %s and response: %s",
			res.StatusCode, method, url, string(responseBody),
		)
		return string(responseBody), err
	}
	return string(responseBody), nil
}

func debugRequest(method string, url string, payload string) {
	if config.App.VeryVeryVerbose {
		if len(payload) > 300 {
			payload = payload[:300] + "(...)"
		}
		fmt.Printf("Method: %s, URL: %s, Payload: %s\n", method, url, payload)
	}
}

func startDevContainers(env Environment, repository string) error {
	jsonData := map[string]string{
		"environment_name": env.Name,
		"repository":       repository,
	}
	if config.App.VeryVerbose {
		println("Start development POST : " + env.GetOrchestratorApi() + "/start_development with name " + env.Name + " and repository " + repository)
	}
	jsonValue, _ := json.Marshal(jsonData)
	response, err := http.Post(
		env.GetOrchestratorApi()+"/start_development",
		"application/json",
		bytes.NewBuffer(jsonValue),
	)
	if err != nil {
		bodyString := ""
		if response != nil && response.Body != nil {
			bodyBytes, _ := io.ReadAll(io.Reader(response.Body))
			bodyString = string(bodyBytes)
			response.Body.Close()
		}
		return fmt.Errorf("error: %v, response: %s", err, bodyString)
	}
	defer response.Body.Close()

	return nil
}
