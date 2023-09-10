package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"src/config"
	"strings"
	"time"

	"github.com/confetti-framework/framework/inter"
)

type token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IdToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

var currentToken = &token{}

func GetAccessToken(cli inter.Cli) (string, error) {
	if currentToken != nil {
		err := FetchNewAccessToken(cli)
		if err != nil {
			return "", err
		}
	}
	return currentToken.AccessToken, nil
}

func FetchNewAccessToken(cli inter.Cli) error {
	if currentToken.AccessToken == "" {
		token, err := getRefreshToken(cli)
		if err != nil {
			return err
		}
		currentToken = token
	}

	return nil
}

func getRefreshToken(cli inter.Cli) (*token, error) {
	url := "https://" + config.Auth0.Domain + "/oauth/device/code"

	payload := strings.NewReader("client_id=" + config.Auth0.ClientId + "&scope=offline_access openid&audience=" + config.Auth0.Audience)

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	content := struct {
		Url        string `json:"verification_uri_complete"`
		DeviceCode string `json:"device_code"`
	}{}
	err = json.Unmarshal(body, &content)
	if err != nil {
		return nil, err
	}

	cli.Comment("\nYou are not logged in here. Go to the following URL to log in:\n")
	cli.Line(content.Url + "\n")
	time.Sleep(5 * time.Second)

	fmt.Print("\u001B[30;1m")
	token, err := getTokenByDeviceCode(cli, content.DeviceCode)
	if err != nil {
		return nil, err
	}
//	println("token.AccessToken:")
//	println(token.AccessToken)
	return token, nil
}

func getTokenByDeviceCode(cli inter.Cli, deviceCode string) (*token, error) {
	url := "https://" + config.Auth0.Domain + "/oauth/token"

	payload := strings.NewReader("grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Adevice_code&device_code=" + deviceCode + "&client_id=" + config.Auth0.ClientId)

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == 403 {
		// Rate limit is 5 seconds
		for i := 5; i >= 1; i-- {
			fmt.Printf("\rRetry in%2d seconds ", i)
			time.Sleep(time.Second)
		}
		return getTokenByDeviceCode(cli, deviceCode)
	}
	fmt.Printf("\r")
	cli.Info("You have successfully logged in")

	content := &token{}

	err = json.Unmarshal(body, content)
	if err != nil {
		return nil, err
	}

	return content, nil
}
