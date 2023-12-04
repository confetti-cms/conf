package services

import (
	"github.com/confetti-framework/framework/inter"
	"src/config"
	"time"
)

var resourceMayHaveChanged bool

func ResourceMayHaveChanged() {
	println("----- Set resource may have changed!")
	resourceMayHaveChanged = true
}

func KeepClientResourcesInSync(cli inter.Cli, env Environment, repo string, since time.Time) error {
	// Generate the component resource files
	err := GenerateComponentFiles(cli, env, repo)
	if err != nil {
		return err
	}
	// If reset, we need to remove all local files before we place the new files
	if since.IsZero() {
		if config.App.Debug {
			println("Removing all local files due to reset")
		}
		err = RemoveAllClientResources()
		if err != nil {
			return err
		}
	}
	err = FetchResources(cli, env, repo, since)
	if err != nil {
		return err
	}
	// Keep the client in sync by running a background job to check on changes	
	go syncClientResources(cli, env, repo, since)
	return nil
}

func syncClientResources(cli inter.Cli, env Environment, repo string, latestCheck time.Time) {
	newLatestCheck := time.Now()
	if !resourceMayHaveChanged {
		time.Sleep(500 * time.Millisecond)
		syncClientResources(cli, env, repo, newLatestCheck)
		return
	}
	// Reset the flag so the resources won't be unnecessarily re-synced
	resourceMayHaveChanged = false
	if config.App.Debug {
		println("----- Resources may have changed: " + latestCheck.Format(time.RFC3339))
	}
	err := FetchResources(cli, env, repo, latestCheck)
	if err != nil {
		cli.Error("Error when fetching client resources: " + err.Error())
	}
	time.Sleep(1 * time.Second)
	syncClientResources(cli, env, repo, newLatestCheck)
}
