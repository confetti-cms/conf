package commands

import (
	"fmt"
	"src/app/services"
	"src/config"
	"time"

	"github.com/confetti-framework/framework/inter"
	"github.com/jedib0t/go-pretty/v6/table"
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
		fmt.Printf("We couldn't find the environment by the query.\nPlease select an environment so we can check if you are authorized.\n")
		runningEnv, err = services.GetEnvironmentByInput(c, "")
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
	containers, chosenOrg := FilterOrganisation(c, containers)
	containers, chosenRepo := FilterRepository(c, containers)

	fmt.Printf("\n\033[0mconf container:query --e=\"%s\" --o=\"%s\" --r=\"%s\"\n\033[0m", env, chosenOrg, chosenRepo)
	renderContainerTable(c, containers)

	printEasterEgg()

	// The watch is preventing the code from ever getting here
	return inter.Success
}

const eggPosition = 36

func printEasterEgg() {
	fmt.Printf("\n\n\n\n\n\n")
	printEmoji(0, " ")
	fmt.Printf("\033[%dD", eggPosition) // Cursor back to the left
	time.Sleep(4000 * time.Millisecond)
	printEmoji(0, "✋")
	printEmoji(0, "👋")
	printEmoji(0, "✋")
	printEmoji(0, "👋")
	printEmoji(0, "✋")
	printEmoji(0, "👋")
	printEmoji(0, "✋")
	printEmoji(0, "👋")
	printEmoji(0, "✋")
	printEmoji(0, "👋")
	printEmoji(0, "✋")
	printEmoji(0, "👋")
	printEmoji(0, "👉")
	printEmoji(1, "👉")
	printEmoji(2, "👉")
	printEmoji(3, "👉")
	printEmoji(4, "👉")
	printEmoji(5, "👉")
	printEmoji(6, "👉")
	printEmoji(7, "👉")
	printEmoji(8, "👉")
	printEmoji(9, "👉")
	printEmoji(10, "👉")
	printEmoji(12, "👉")
	printEmoji(13, "👉")
	printEmoji(14, "👉")
	printEmoji(15, "👉")
	printEmoji(16, "")

	typeMachine(18, "You found a hidden egg today,\n")
	typeMachine(17, "A glimpse of what's on the way.\n\n")
	typeMachine(17, "Logs and containers, all in one,\n")
	typeMachine(17, "Across projects, work gets done.\n\n")
	typeMachine(18, "Confetti runs like it's local,\n")
	typeMachine(19, "Fast and clean, fully vocal.\n\n")

	time.Sleep(500 * time.Millisecond)
	fmt.Print("\n")
	time.Sleep(200 * time.Millisecond)
	fmt.Print("\n")
	time.Sleep(200 * time.Millisecond)
	fmt.Print("\n")
	time.Sleep(400 * time.Millisecond)
	fmt.Print("\n")
	time.Sleep(1000 * time.Millisecond)
	fmt.Print("\n")
	time.Sleep(2000 * time.Millisecond)
	fmt.Print("\n")
}

func typeMachine(prefix int, line string) {
	printCharacter(prefix, " ", 0)
	for _, char := range line {
		fmt.Printf("%c", char)
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Print("\n")
}

func printEmoji(pos int, e string) {
	fmt.Printf("                                                  \r") // Remove previous emoji
	// if pos > 0 {
	// fmt.Printf("\033[%dC", pos) // Cursor to the right
	// }
	printCharacter(15+pos, " ", 0) // Fill the rest with spaces
	fmt.Printf("%s", e)            // Print emoji

	// Print 🥚 at the right position (always with 15 white spaces from the very left - position of the hand
	printCharacter(15-pos, " ", 0) // Fill the rest with spaces
	if pos <= 15 {
		fmt.Print("🥚")
		fmt.Printf("\033[%dD", eggPosition) // Cursor back to the left
	} else {
		fmt.Print("🐣\n")
		time.Sleep(2000 * time.Millisecond)
		fmt.Print("\n")
		time.Sleep(1000 * time.Millisecond)
		fmt.Print("\n")
		time.Sleep(400 * time.Millisecond)
		fmt.Print("\n")
		time.Sleep(200 * time.Millisecond)
		fmt.Print("\n")
		time.Sleep(200 * time.Millisecond)
		fmt.Print("\n")
		time.Sleep(100 * time.Millisecond)
		fmt.Print("\n")
	}

	if pos == 0 {
		time.Sleep(300 * time.Millisecond)
	} else {
		time.Sleep(120 * time.Millisecond)
	}
	fmt.Printf("\033[%dD", eggPosition) // Cursor back to the left
}

func printCharacter(times int, char string, duration time.Duration) {
	for i := 0; i < times; i++ {
		fmt.Print(char)
		time.Sleep(duration)
	}
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
	if len(names) == 1 {
		return "", fmt.Errorf("no environments found, please define an environment in your config.json5 file")
	}
	envName = c.Choice("You can narrow down your search by selecting an environment:", names...)
	if envName == AllEnvironments {
		envName = "*"
	}

	return envName, nil
}

const AllOrganisations = "all organisations"

func FilterOrganisation(c inter.Cli, containers []services.ContainerInformation) ([]services.ContainerInformation, string) {
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
		return containers, orgNames[0]
	}

	// If there are no organisations, we return an empty slice
	if len(orgNames) == 0 {
		return []services.ContainerInformation{}, "*"
	}

	// Multiple organisations, prompt user
	// If the user has specified an organisation, we already have the containers filtered
	choices := append([]string{AllOrganisations}, orgNames...)
	selected := c.Choice("Select an organisation:", choices...)
	if selected == AllOrganisations {
		return containers, "*"
	}
	filtered := []services.ContainerInformation{}
	for _, container := range containers {
		if container.UmbrellaOrganization == selected {
			filtered = append(filtered, container)
		}
	}
	return filtered, selected
}

const AllRepositories = "all repositories"

func FilterRepository(c inter.Cli, containers []services.ContainerInformation) ([]services.ContainerInformation, string) {
	// Get all unique repository names from the containers
	repoMap := map[string]bool{}
	for _, container := range containers {
		if container.UmbrellaRepository == "" {
			panic(fmt.Sprintf("Container %s has no umbrella repository set, this should not happen. Locator: %s", container.Name, container.Locator))
		}
		repoMap[container.UmbrellaRepository] = true
	}
	repoNames := []string{}
	for repoName := range repoMap {
		repoNames = append(repoNames, repoName)
	}

	// If there is only one repository, we return all containers
	if len(repoNames) == 1 {
		return containers, repoNames[0]
	}

	// If there are no repositories, we return an empty slice
	if len(repoNames) == 0 {
		return []services.ContainerInformation{}, "*"
	}

	// Multiple repositories, prompt user
	choices := append([]string{AllRepositories}, repoNames...)
	selected := c.Choice("Select a repository:", choices...)
	if selected == AllRepositories {
		return containers, "*"
	}
	filtered := []services.ContainerInformation{}
	for _, container := range containers {
		if container.UmbrellaRepository == selected {
			filtered = append(filtered, container)
		}
	}
	return filtered, selected
}

func renderContainerTable(c inter.Cli, containers []services.ContainerInformation) {
	ta := c.Table()
	ta.AppendHeader(table.Row{"Name", "target", "status"})
	for _, container := range containers {
		statusColor := "\033[32m" // green
		switch container.Status {
		case "running":
			statusColor = "\033[32m" // green
		case "stopped", "exited":
			statusColor = "\033[31m" // red
		case "paused":
			statusColor = "\033[33m" // yellow
		default:
			statusColor = "\033[36m" // cyan
		}
		ta.AppendRow(table.Row{
			fmt.Sprintf("\033[34m%s\033[0m", container.Name),   // blue
			fmt.Sprintf("\033[34m%s\033[0m", container.Target), // magenta
			fmt.Sprintf("%s%s\033[0m", statusColor, container.Status),
		})
	}
	ta.Render()
}
