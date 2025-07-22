package commands

import (
	"fmt"
	"src/app/services"
	"src/config"

	"github.com/confetti-framework/framework/inter"
)

type ContainerQuery struct {
	Directory       string `short:"d" flag:"directory" description:"Root directory of the project, defaults to the current directory"`
	Environment     string `short:"e" flag:"environment" description:"The environment name in the config.json5 file"`
	Organisation    string `short:"o" flag:"organisation" description:"The organisation name (e.g. agency-name is the organisation when repository is: 'agency-name/website-name')"`
	Repository      string `short:"r" flag:"repository" description:"The repository name (e.g. website-name is the repository when repository is: 'agency-name/website-name')"`
	Verbose         bool   `short:"v" description:"Show events"`
	VeryVerbose     bool   `short:"vv" description:"Show more events"`
	VeryVeryVerbose bool   `short:"vvv" description:"Show all events"`
	Local           bool   `short:"l" description:"Only used by Confetti developers to run against the local orchestrator"`
}

func (l ContainerQuery) Name() string {
	return "container:query"
}

func (l ContainerQuery) Description() string {
	return "Query for containers and get more information about them"
}

func (l ContainerQuery) Handle(c inter.Cli) inter.ExitCode {
	config.App.Verbose = l.Verbose || l.VeryVerbose || l.VeryVeryVerbose
	config.App.VeryVerbose = l.VeryVerbose || l.VeryVeryVerbose
	config.App.VeryVeryVerbose = l.VeryVeryVerbose
	config.App.Local = l.Local
	root, err := getDirectoryOrCurrent(l.Directory)
	if err != nil {
		c.Error(err.Error())
		return inter.Failure
	}
	config.Path.Root = root

	if config.App.Verbose {
		c.Info("Use directory: %s", root)
	}

	// Get Environment for querying containers
	env := l.Environment
	if env == "" {
		fmt.Println("\n\033[34mConfetti container:log\n\033[0m")
		env, err = GetEnvironmentName(c, l.Environment)
		if err != nil {
			c.Error(fmt.Sprintf("Error getting environment: %s", err))
			return inter.Failure
		}
	}

	// Get the running environment, only to validate the user is authorized
	runningEnv, err := services.GetEnvironmentByInput(c, env)
	if err != nil {
		// If we can't get the environment by the query env (e.g. "all environments"),
		// We ask the user to select an running environment
		fmt.Printf("We couldn't find the environment by the query env.\nPlease select an environment so we can check if you are authorized.\n")
		runningEnv, err := services.GetEnvironmentByInput(c, "")
		if err != nil {
			c.Error(fmt.Sprintf("Error getting environment: %s", err))
			return inter.Failure
		}
	}

	// Get all basic information about the containers the containers
	query := services.QueryContainerOptions{
		Environment:          env,
		UmbrellaOrganization: l.Organisation,
		UmbrellaRepository:   l.Repository,
	}
	containers, err := services.GetContainers(c, runningEnv, query)
	// Get the organisation name for the query
	if err != nil {
		c.Error(fmt.Sprintf("Error getting containers: %s", err))
		return inter.Failure
	}
	containers = FilterOrganisation(c, containers)

	// for now dump the containers to debug
	for _, container := range containers {
		fmt.Printf("\n\033[34mContainer: %s\n\033[0m", container.Name)
		fmt.Printf("  Locator: %s\n", container.Locator)
		

	fmt.Printf("\n\033[34mconf container:query --env=\"%s\"\n\033[0m", env)

	// The watch is preventing the code from ever getting here
	return inter.Success
}

const AllEnvironments = "all environments"

func GetEnvironmentName(c inter.Cli, envName string) (string, error) {
	appConfig, err := services.GetAppConfig()
	if err != nil {
		return "", err
	}
	names := []string{AllEnvironments}
	for _, environment := range appConfig.Environments {
		names = append(names, environment.Name)
	}
	if envName == AllEnvironments {
		envName = "*"
	}

	return envName, nil
}

func FilterOrganisation(c inter.Cli, containers []services.ContainerInfo) []services.ContainerInfo {
	// Get all unique organisation names from the containers
	orgMap := map[string]bool{}
	for _, container := range containers {
		if container.UmbrellaOrganization == "" {
			panic(fmt.Sprintf("Container %s has no umbrella organization set, this should not happen. Locator: %s", container.Name, container.Locator))
		}
		orgMap[container.UmbrellaOrganization] = true
	}
	orgNames := []string{}
	for orgName := range orgMap {
		orgNames = append(orgNames, orgName)
	}

	// If there is only one organisation, we return all containers
	if len(orgNames) == 1 {
		return containers
	}

	// If there are no organisations, we return an empty slice
	if len(orgNames) == 0 {
		return []services.ContainerInfo{}
	}

	// Multiple organisations, prompt user
	// If the user has specified an organisation, we already have the containers filtered
	choices := append([]string{"*"}, orgNames...)
	selected := c.Choice("Select an organisation:", choices...)
	if selected == "*" {
		return containers
	}
	filtered := []services.ContainerInfo{}
	for _, container := range containers {
		if container.UmbrellaOrganization == selected {
			filtered = append(filtered, container)
		}
	}
	return filtered
}
