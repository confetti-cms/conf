package services

import (
	"src/config"
	"time"

	"github.com/confetti-framework/framework/inter"
)

var resourceMayHaveChanged bool

func ResourceMayHaveChanged() {
	if config.App.VeryVerbose {
		println("Resource may have changed")
	}
	resourceMayHaveChanged = true
}

func UpdateComponents(cli inter.Cli, env Environment, repo string, since time.Time, reset bool) error {
	// If reset, we need to remove all local files before we place the new files
	if reset {
		if config.App.VeryVerbose {
			println("Removing all local files due to reset")
		}
		err := RemoveAllLocalResources()
		if err != nil {
			return err
		}
	}
	go keepLocalResourcesInSync(cli, env, repo, since)
	return nil
}

func keepLocalResourcesInSync(cli inter.Cli, env Environment, repo string, checkSince time.Time) {
	ResourceMayHaveChanged()
	// To prevent that checkSince is the same as newCheckSince, we wait one second.
	time.Sleep(time.Second)
	// Keep the client in sync by running a background job to check on changes
	go syncClientResources(cli, env, repo, checkSince)
}

func syncClientResources(cli inter.Cli, env Environment, repo string, checkSince time.Time) {
	newCheckSince := time.Now()
	if !resourceMayHaveChanged {
		time.Sleep(200 * time.Millisecond)
		syncClientResources(cli, env, repo, checkSince)
		return
	}
	// Reset the flag so the resources won't be unnecessarily re-synced
	resourceMayHaveChanged = false
	if config.App.VeryVerbose {
		println("Resources may have changed: " + checkSince.Format(time.RFC3339))
	}
	for range 10 {
		newCheckSince = time.Now()
		err := FetchResources(cli, env, repo, checkSince)
		if err != nil {
			if config.App.VeryVerbose {
				println("Error when fetching client resources (Retrying after 3 seconds...): " + err.Error())
			}
			time.Sleep(3 * time.Second)
			if config.App.VeryVerbose {
				println("Retrying...")
			}

			err := FetchResources(cli, env, repo, checkSince)
			if err != nil {
				cli.Error("Error when fetching client resources (second time): " + err.Error())
				return
			}
		}
		time.Sleep(1 * time.Second)
		checkSince = newCheckSince
	}
	syncClientResources(cli, env, repo, newCheckSince)
}
