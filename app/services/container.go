package services

import (
	"encoding/json"
	"net/http"
	"net/url"
	"src/config"

	"github.com/confetti-framework/framework/inter"
)

type ContainerInformation struct {
}

type ContainerInfo struct {
	ID                   string                 `json:"id"`
	Locator              string                 `json:"locator"`
	SourceOrganization   string                 `json:"source_organization"`
	SourceRepository     string                 `json:"source_repository"`
	UmbrellaOrganization string                 `json:"umbrella_organization"`
	UmbrellaRepository   string                 `json:"umbrella_repository"`
	Name                 string                 `json:"name"`
	Target               string                 `json:"target"`
	Status               string                 `json:"status"`
	Ports                []string               `json:"ports"`
	NetworkName          string                 `json:"network_name"`
	Environment          EnvironmentInformation `json:"environment"`
}

type EnvironmentInformation struct {
	Name  string `json:"name"`
	Stage string `json:"stage"`
}

type QueryContainerOptions struct {
	Environment          string `json:"environment"`
	UmbrellaOrganization string `json:"umbrella_organization"`
	UmbrellaRepository   string `json:"umbrella_repository"`
}

func GetContainers(cli inter.Cli, runningEnv Environment, options QueryContainerOptions) ([]ContainerInfo, error) {
	baseUrl := GetOrchestratorContainerListUrl()
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	if options.Environment != "" {
		q.Set("environment", options.Environment)
	}
	if options.UmbrellaOrganization != "" {
		q.Set("umbrella_organization", options.UmbrellaOrganization)
	}
	if options.UmbrellaRepository != "" {
		q.Set("umbrella_repository", options.UmbrellaRepository)
	}
	u.RawQuery = q.Encode()

	resp, err := Send(cli, u.String(), nil, http.MethodPut, runningEnv, "")
	if err != nil {
		return nil, err
	}

	type responseData struct {
		Data []ContainerInfo `json:"data"`
	}
	var result responseData
	err = json.Unmarshal([]byte(resp), &result)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func GetOrchestratorContainerListUrl() string {
	if config.App.Local {
		return OrchestratorApiLocalhost + "/container_list"
	}
	return OrchestratorApiDefault + "/container_list"
}
