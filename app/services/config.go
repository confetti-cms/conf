package services

import (
	"fmt"

	"path/filepath"
	"src/config"
	"strings"

	"github.com/confetti-framework/framework/inter"

	"os"

	"github.com/titanous/json5"
)

const configFile = "app_config.json5"

type Hosts []string
type Paths []string

type ContainerConfig struct {
	Name  string `json:"name"`
	Hosts Hosts  `json:"hosts"`
	Paths Paths  `json:"paths"`
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

const OrchestratorApiLocalhost = "http://api.confetti-cms.localhost/orchestrator"
const OrchestratorApiDefault = "https://api.confetti-cms.com/orchestrator"

type Environment struct {
	Name       string            `json:"name"`
	Local      bool              `json:"local"`
	Containers []ContainerConfig `json:"containers"`
}

func (e Environment) GetOrchestratorApi() string {
	if e.Local {
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
			break
		}
	}
	// Find match
	for _, container := range e.Containers {
		if container.Name == service {
			match = container
			break
		}
	}
	host := match.Hosts[0]
	method := "https://"
	if e.Local {
		method = "http://"
	}
	return method + host + getUriByAlias(match, service)
}

func getUriByAlias(cConfig ContainerConfig, service string) string {
	uri := ""
	if len(cConfig.Paths) > 0 {
		// For now, we only support 1 path max
		uri += "/" + strings.TrimLeft(cConfig.Paths[0], "/")
	}
	// Replace __SERVICE__ with the service name
	uri = strings.ReplaceAll(uri, "__SERVICE__", service)

	return strings.TrimRight(uri, "/")
}

type AppConfig struct {
	Environments []Environment `json:"environments"`
}

func GetAppConfig() (AppConfig, error) {
	aConfig := AppConfig{}
	content, err := os.ReadFile(filepath.Join(config.Path.Root, configFile))
	if err != nil {
		return aConfig, fmt.Errorf("probably, you are not running this command in a Confetti project. Error: %s", err)
	}

	err = json5.Unmarshal(content, &aConfig)
	if err != nil {
		return aConfig, fmt.Errorf("invalid JSON5 content in %s: %s", configFile, err)
	}

	return aConfig, nil
}

func GetEnvironmentByInput(c inter.Cli, envName string) (Environment, error) {
	appConfig, err := GetAppConfig()
	if err != nil {
		return Environment{}, err
	}
	names := []string{}
	for _, environment := range appConfig.Environments {
		names = append(names, environment.Name)
	}
	if len(names) == 1 {
		return appConfig.Environments[0], nil
	}
	if envName == "" {
		envName = c.Choice("Choose your environment", names...)
	}
	for _, environment := range appConfig.Environments {
		if environment.Name == envName {
			if config.App.VeryVerbose {
				fmt.Println("Environment name is:", envName)
			}
			return environment, nil
		}
	}

	return Environment{}, fmt.Errorf("the name %s does not match any environment. Available names are %s", envName, strings.Join(names, ", "))
}
