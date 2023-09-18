package services

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
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

func getUriByAlias(config ContainerConfig, service string) string {
	uri := ""
	if len(config.Paths) > 0 {
		// For now, we only support 1 path max
		uri += "/" + strings.TrimLeft(config.Paths[0], "/")
	}
	if config.UserServiceInUri {
		// For now, we only support 1 suffix path max
		uri += "/" + service
	}

	return strings.TrimRight(uri, "/")
}

type AppConfig struct {
	Environments []Environment `json:"environments"`
}

func GetAppConfig(dir string) (AppConfig, error) {
	config := AppConfig{}
	content, err := ioutil.ReadFile(filepath.Join(dir, configFile))
	if err != nil {
		return config, fmt.Errorf("error reading file: %s", err)
	}

	err = json5.Unmarshal(content, &config)
	if err != nil {
		return config, fmt.Errorf("error unmarshal json5: %s", err)
	}

	return config, nil
}

func GetEnvironmentByInput(c inter.Cli, dir string) (Environment, error) {
	config, err := GetAppConfig(dir)
	if err != nil {
		return Environment{}, err
	}
	keys := []string{}
	for _, environment := range config.Environments {
		keys = append(keys, environment.Key)
	}
	if len(keys) == 1 {
		return config.Environments[0], nil
	}
	envKey := c.Choice("Choose your environment", keys...)
	for _, environment := range config.Environments {
		if environment.Key == envKey {
			return environment, nil
		}
	}

	return Environment{}, errors.New("The key " + envKey + " does not match any environment")
}
