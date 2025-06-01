package services

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
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

var currentToken *token

func GetAccessToken(cli inter.Cli, env Environment) (string, error) {
	if currentToken == nil {
		err := tryTokenFromFile(cli, env)
		if err != nil {
			return "", err
		}
	}
	return currentToken.AccessToken, nil
}

func tryTokenFromFile(cli inter.Cli, env Environment) error {
	// Check if the file exists
	_, err := os.Stat(path.Join(config.Path.Root, sharedResourcesDir, "auth_token.json"))
	if os.IsNotExist(err) {
		if config.App.Verbose {
			fmt.Println("Token file does not exist, creating a new one...")
		}
		err := createOrUpdateAuthTokenFile(cli)
		if err != nil {
			return fmt.Errorf("unable to decode current token: %w", err)
		}
	}
	err = useCurrentTokenFromFile()
	if err != nil {
		return fmt.Errorf("error using current token from file: %w", err)
	}
	valid, err := currentTokenIsValid(cli, env)
	if err != nil {
		return fmt.Errorf("error checking if current token is valid: %w", err)
	}

	// If the token is not valid, create a new one
	if !valid {
		err := createOrUpdateAuthTokenFile(cli)
		if err != nil {
			return fmt.Errorf("token from file not valid, unable to decode current token: %w", err)
		}
	}

	return nil
}

const getRolesEndpoint = "/users/me"

func currentTokenIsValid(cli inter.Cli, env Environment) (bool, error) {
	// Create request
	url := env.GetServiceUrl("confetti-cms/auth") + getRolesEndpoint
	req, err := http.NewRequest(
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return false, fmt.Errorf("unable to create request: %v", err)
	}
	h := req.Header
	h.Add("Content-Type", "application/json")
	h.Add("Authorization", "Bearer "+currentToken.AccessToken)
	req.Header = h
	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return false, nil
	}
	if resp.StatusCode >= 300 {

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, fmt.Errorf("error reading response body: %v", err)
		}
		bodyString := string(bodyBytes)
		cli.Error("Unsuccessful response status: %d: Response body: %s", resp.StatusCode, bodyString)
		return false, fmt.Errorf("unsuccessful response status: %d. Response body: %s", resp.StatusCode, bodyString)
	}
	return true, nil
}

func useCurrentTokenFromFile() error {
	// The file exists, decode it
	file, err := os.Open(path.Join(config.Path.Root, sharedResourcesDir, "auth_token.json"))
	if err != nil {
		return fmt.Errorf("unable to open auth_token.json file: %w", err)
	}
	defer file.Close()
	if currentToken == nil {
		currentToken = &token{}
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(currentToken)
	if err != nil {
		return fmt.Errorf("unable to decode current token: %w", err)
	}
	return nil
}

func createOrUpdateAuthTokenFile(cli inter.Cli) error {
	// The directory doesn't exist, create it
	err := os.MkdirAll(path.Join(config.Path.Root, sharedResourcesDir), os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create directory: %w", err)
	}
	// Generate the token
	err = FetchNewAccessToken(cli)
	if err != nil {
		return fmt.Errorf("unable to fetch new access token: %w", err)
	}
	// Open file
	file, err := os.OpenFile(path.Join(config.Path.Root, sharedResourcesDir, "auth_token.json"), os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to open or create file: %w", err)
	}
	defer file.Close()
	// Save the token to the file
	token, err := json.Marshal(currentToken)
	if err != nil {
		return fmt.Errorf("unable to marshal token: %w", err)
	}
	_, err = file.Write(token)
	if err != nil {
		return fmt.Errorf("unable to write to file: %w", err)
	}
	if config.App.Verbose {
		fmt.Printf("Token file created: %s\n", file.Name())
	}

	return nil
}

func FetchNewAccessToken(cli inter.Cli) error {
	token, err := getRefreshToken(cli)
	if err != nil {
		return err
	}
	currentToken = token

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

	// Remove the last line of the screen
	fmt.Printf("\r                                                                      \n")
	cli.Comment("Login to sync your local code with the server")
	// If windows, the user need to give access to open port 8001
	if runtime.GOOS == "windows" {
		cli.Comment("And allow access to port 8001 for hot reload to work")
	}

	cli.Comment("\n\033[34m               ╭──────────────────────╮\n               │ \033[0mPress enter to login \033[34m│\n               ╰──────────────────────╯\n")

	// When project already init, do not wait for the enter key press
	_, err = os.Stat(filepath.Join(config.Path.Root, "vendor"))
	if os.IsNotExist(err) {
		buf := bufio.NewReader(os.Stdin)
		_, _ = buf.ReadBytes('\n') // Wait for the key press
	}
	err = OpenUrl(content.Url)
	if err != nil {
		return nil, err
	}
	// Clean entire screen
	print("\033[H\033[2J")
	cli.Comment("Waiting...")
	time.Sleep(5 * time.Second)

	fmt.Print("\u001B[30;1m")
	token, err := getTokenByDeviceCode(cli, content.DeviceCode)
	if err != nil {
		return nil, err
	}
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

	if res.StatusCode == http.StatusForbidden {
		// Rate limit is 5 seconds
		for i := 5; i >= 1; i-- {
			fmt.Printf("\rRetry in%2d seconds ", i)
			time.Sleep(time.Second)
		}
		return getTokenByDeviceCode(cli, deviceCode)
	}

	// Clean entire screen
	print("\033[H\033[2J")
	cli.Info("Welcome back! You’re logged in 🥳")
	cli.Line("Syncing your local code with the server...")

	content := &token{}

	err = json.Unmarshal(body, content)
	if err != nil {
		return nil, err
	}

	return content, nil
}
