package services

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/confetti-framework/framework/inter"
)

type ContainerInformation struct {
	ID                   string                 `json:"id"`
	Locator              string                 `json:"locator"`
	SourceOrganization   string                 `json:"source_organization"`
	SourceRepository     string                 `json:"source_repository"`
	UmbrellaOrganization string                 `json:"umbrella_organization"`
	UmbrellaRepository   string                 `json:"umbrella_repository"`
	Name                 string                 `json:"name"`
	Target               string                 `json:"target"`
	Status               string                 `json:"status"`
	Ports                []uint                 `json:"ports"`
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

func GetContainers(cli inter.Cli, runningEnv Environment, options QueryContainerOptions) ([]ContainerInformation, error) {
	u, err := url.Parse(GetOrchestratorContainerListUrl(runningEnv.Local))
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

	resp, err := Send(cli, u.String(), nil, http.MethodGet, runningEnv, "", 30*time.Second)
	if err != nil {
		return nil, err
	}

	type responseData struct {
		Data []ContainerInformation `json:"data"`
	}
	var result responseData
	err = json.Unmarshal([]byte(resp), &result)
	if err != nil {
		return nil, err
	}

	// For demo, we add a dummy data with unique locators
	result.Data = append(result.Data,
		// DEV
		ContainerInformation{
			ID:                   "dev-website-id",
			Locator:              "flip-agency/potlodenshop/website?env=dev",
			SourceOrganization:   "flip-agency",
			SourceRepository:     "potlodenshop",
			UmbrellaOrganization: "flip-agency",
			UmbrellaRepository:   "potlodenshop",
			Name:                 "website",
			Target:               "CMD",
			Status:               "running",
			Ports:                []uint{8080, 443},
			NetworkName:          "dummy-network",
			Environment: EnvironmentInformation{
				Name:  "dev",
				Stage: "development",
			},
		}, ContainerInformation{
			ID:                   "dev-sharpening-id",
			Locator:              "flip-agency/potlodenshop/sharpening-service?env=dev",
			SourceOrganization:   "flip-agency",
			SourceRepository:     "potlodenshop",
			UmbrellaOrganization: "flip-agency",
			UmbrellaRepository:   "potlodenshop",
			Name:                 "sharpening-service",
			Target:               "CMD",
			Status:               "stopped",
			Ports:                []uint{80},
			NetworkName:          "dummy-network",
			Environment: EnvironmentInformation{
				Name:  "dev",
				Stage: "development",
			},
		},
		// PROD
		ContainerInformation{
			ID:                   "prod-website-id",
			Locator:              "flip-agency/potlodenshop/website?env=prod",
			SourceOrganization:   "flip-agency",
			SourceRepository:     "potlodenshop",
			UmbrellaOrganization: "flip-agency",
			UmbrellaRepository:   "potlodenshop",
			Name:                 "website",
			Target:               "CMD",
			Status:               "running",
			Ports:                []uint{8080, 443},
			NetworkName:          "dummy-network",
			Environment: EnvironmentInformation{
				Name:  "prod",
				Stage: "production",
			},
		},
		ContainerInformation{
			ID:                   "prod-sharpening-id",
			Locator:              "flip-agency/potlodenshop/sharpening-service?env=prod",
			SourceOrganization:   "flip-agency",
			SourceRepository:     "potlodenshop",
			UmbrellaOrganization: "flip-agency",
			UmbrellaRepository:   "potlodenshop",
			Name:                 "sharpening-service",
			Target:               "CMD",
			Status:               "running",
			Ports:                []uint{80},
			NetworkName:          "dummy-network",
			Environment: EnvironmentInformation{
				Name:  "prod",
				Stage: "production",
			},
		},
	)

	return result.Data, nil
}

func GetOrchestratorContainerListUrl(isLocal bool) string {
	if isLocal {
		return OrchestratorApiLocalhost + "/container_list"
	}
	return OrchestratorApiDefault + "/container_list"
}
