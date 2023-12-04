package services

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"src/config"
	"strings"

	"github.com/confetti-framework/framework/inter"

	"github.com/titanous/json5"
)

const configFile = "app_config.json5"

type Hosts []string
type Paths []string

type ContainerConfig struct {
	Name             string `json:"name"`
	Hosts            Hosts  `json:"hosts"`
	Paths            Paths  `json:"paths"`
	UserServiceInUri bool   `json:"user_service_in_uri"`
}

func (c *ContainerConfig) GetAllURLCombinations(defaultUri string) []string {
	var combinations []string

	for _, host := range c.Hosts {
		uri := defaultUri
		if len(c.Paths) > 0 {
			// For now, we only support 1 path max
			uri = strings.TrimLeft(c.Paths[0], "/")
		}
		url := host + "/" + uri
		combinations = append(combinations, strings.TrimRight(url, "/"))
	}

	return combinations
}

const OrchestratorApiDefault = "http://api.confetti-cms.com/orchestrator"
const OrchestratorApiLocalhost = "http://api.confetti-cms.localhost/orchestrator"

type Environment struct {
	Key            string            `json:"key"`
	RunOnLocalhost bool              `json:"run_on_localhost"`
	Containers     []ContainerConfig `json:"containers"`
}

func (e Environment) GetOrchestratorApi() string {
	if e.RunOnLocalhost {
		return OrchestratorApiLocalhost
	}
	return OrchestratorApiDefault
}

func (e Environment) GetAllHosts() []string {
	hosts := []string{}
	hostMap := make(map[string]bool)
	for _, container := range e.Containers {
		for _, host := range container.Hosts {
			if !hostMap[host] {
				hostMap[host] = true
				hosts = append(hosts, host)
			}
		}
	}
	return hosts
}

func (e Environment) GetExplicitHosts() []string {
	hosts := []string{}
	hostMap := make(map[string]bool)
	for _, container := range e.Containers {
		for _, host := range container.Hosts {
			if container.Name == "" {
				continue
			}
			if !hostMap[host] {
				hostMap[host] = true
				hosts = append(hosts, host)
			}
		}
	}
	return hosts
}

func (e Environment) GetServiceUrl(service string) string {
	match := ContainerConfig{}
	// Set default
	for _, container := range e.Containers {
		if container.Name == "" {
			match = container
		}
	}
	// Find match
	for _, container := range e.Containers {
		if container.Name == service {
			match = container
		}
	}
	host := match.Hosts[0]
	return "http://" + host + getUriByAlias(match, service)
}

func getUriByAlias(cConfig ContainerConfig, service string) string {
	uri := ""
	if len(cConfig.Paths) > 0 {
		// For now, we only support 1 path max
		uri += "/" + strings.TrimLeft(cConfig.Paths[0], "/")
	}
	if cConfig.UserServiceInUri {
		// For now, we only support 1 suffix path max
		uri += "/" + service
	}

	return strings.TrimRight(uri, "/")
}

type AppConfig struct {
	Environments []Environment `json:"environments"`
}

func GetAppConfig() (AppConfig, error) {
	aConfig := AppConfig{}
	content, err := ioutil.ReadFile(filepath.Join(config.Path.Root, configFile))
	if err != nil {
		return aConfig, fmt.Errorf("probably, you are not running this command in a Confetti project. Error: %s", err)
	}

	err = json5.Unmarshal(content, &aConfig)
	if err != nil {
		return aConfig, fmt.Errorf("error unmarshal json5: %s", err)
	}

	return aConfig, nil
}

func GetEnvironmentByInput(c inter.Cli, envKey string) (Environment, error) {
	appConfig, err := GetAppConfig()
	if err != nil {
		return Environment{}, err
	}
	keys := []string{}
	for _, environment := range appConfig.Environments {
		keys = append(keys, environment.Key)
	}
	if len(keys) == 1 {
		return appConfig.Environments[0], nil
	}
	if envKey == "" {
		envKey = c.Choice("Choose your environment", keys...)
	}
	for _, environment := range appConfig.Environments {
		if environment.Key == envKey {
			if config.App.Debug {
				fmt.Println("Environment key is:", envKey)
			}
			return environment, nil
		}
	}

	return Environment{}, fmt.Errorf("the key %s does not match any environment. Available keys are %s", envKey, strings.Join(keys, ", "))
}
