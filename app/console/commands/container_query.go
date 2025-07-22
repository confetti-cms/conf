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

	env := l.Environment
	if env == "" {
		fmt.Println("\n\033[34mConfetti container:log\n\033[0m")
		env, err = GetEnvironmentName(c, l.Environment)
		if err != nil {
			c.Error(fmt.Sprintf("Error getting environment: %s", err))
			return inter.Failure
		}
	}

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
	envName = c.Choice("You can narrow down your search by selecting an environment:", names...)
	if envName == AllEnvironments {
		envName = "*"
	}

	return envName, nil
}
