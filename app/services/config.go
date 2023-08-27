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

type Environment struct {
	Key        string            `json:"key"`
	Containers []ContainerConfig `json:"containers"`
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

func (e Environment) GetServiceUrl(serviceName string) string {
	_default := ContainerConfig{}
	for _, container := range e.Containers {
		if container.Name == serviceName {
			host := container.Hosts[0]
			return "http://" + host + "/" + serviceName
		}
		if container.Name == "" {
			_default = container
		}
	}
	host := _default.Hosts[0]
	return "http://" + host + "/" + serviceName
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
