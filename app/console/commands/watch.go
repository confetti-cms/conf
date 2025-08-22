package commands

import (
	"fmt"
	"src/app/services"
	"src/app/services/event_bus"
	"src/app/services/scanner"
	"src/config"
	"time"

	"github.com/confetti-framework/errors"

	"github.com/confetti-framework/framework/inter"
)

type Watch struct {
	Directory       string `short:"d" flag:"directory" description:"Root directory of the project, defaults to the current directory"`
	Environment     string `short:"e" flag:"environment" description:"The environment name in the config.json5 file, default 'dev'"`
	Reset           bool   `short:"r" flag:"reset" description:"All files are parsed again"`
	Verbose         bool   `short:"v" description:"Show events"`
	VeryVerbose     bool   `short:"vv" description:"Show more events"`
	VeryVeryVerbose bool   `short:"vvv" description:"Show all events"`
}

func (t Watch) Name() string {
	return "watch"
}

func (t Watch) Description() string {
	return "Keeps your local files in sync with the server."
}

func (t Watch) Handle(c inter.Cli) inter.ExitCode {
	updateResourcesSince := time.Time{}
	if !t.Reset {
		updateResourcesSince = time.Now()
	}
	config.App.Verbose = t.Verbose || t.VeryVerbose || t.VeryVeryVerbose
	config.App.VeryVerbose = t.VeryVerbose || t.VeryVeryVerbose
	config.App.VeryVeryVerbose = t.VeryVeryVerbose
	root, err := getDirectoryOrCurrent(t.Directory)
	if err != nil {
		c.Error(err.Error())
		return inter.Failure
	}
	config.Path.Root = root

	if config.App.Verbose {
		c.Info("Use directory: %s", root)
	}
	fmt.Println("\n\033[34mConfetti watch\n\033[0m") // blue
	env, err := services.GetEnvironmentByInput(c, t.Environment)
	if err != nil {
		c.Error(fmt.Sprintf("Error getting environment: %s", err))
		return inter.Failure
	}

	if env.Options.DevTools {
		// Open the event bus server
		if config.App.VeryVerbose {
			fmt.Println("->> Opening the event bus server on http://localhost:8001/messages")
		}
		go event_bus.Publish()
	} else {
		if config.App.VeryVerbose {
			c.Line("Config DevTool is set to false. The event bus server is not started.")
		}
	}

	// Get commit of the remote repository
	remoteCommit := services.GetGitRemoteCommit()
	repo, err := services.GetRepositoryName(root)
	if err != nil {
		c.Error(err.Error())
		return inter.Failure
	}

	// Checkout the repository
	fmt.Printf("Sync files...                                                         ")
	err = services.SendCheckout(c, env, services.CheckoutBody{
		Commit: remoteCommit,
		Reset:  t.Reset,
		Parse:  false,
	}, repo)
	if err != nil {
		c.Error(err.Error())
		if !errors.Is(err, services.UserError) {
			return inter.Failure
		}
	}

	// Apply all local changes
	if config.App.VeryVerbose {
		fmt.Println("->> Applying local changes")
	}
	filesToSync := services.PatchDir(c, env, remoteCommit, c.Writer(), repo)
	// Remove loading bar
	fmt.Printf("\r                                                                      \n")

	// Run composer install locally if vendor directory is missing
	if config.App.VeryVerbose {
		fmt.Println("->> Checking composer dependencies")
	}
	err = services.ComposerInstall(c, env)
	if err != nil {
		c.Error(err.Error())
		if !errors.Is(err, services.UserError) {
			return inter.Failure
		}
	}

	// Parse all base components (other components wil extend this components)
	if config.App.VeryVerbose {
		fmt.Println("->> Parsing base components")
	}
	err = services.ParseBaseComponents(c, env, repo)
	if err != nil {
		c.Error(err.Error())
		if !errors.Is(err, services.UserError) {
			return inter.Failure
		}
	}

	// Parse components
	if config.App.VeryVerbose {
		fmt.Println("->> Parsing components")
	}
	if t.Reset {
		err = services.ParseAllComponents(c, env, repo)
		if err != nil {
			c.Error(err.Error())
			if !errors.Is(err, services.UserError) {
				return inter.Failure
			}
		}
	} else {
		for _, file := range filesToSync {
			err = services.ParseComponent(c, env, services.ParseComponentBody{
				File: file,
			}, repo)
			if err != nil {
				c.Error(err.Error())
				if !errors.Is(err, services.UserError) {
					return inter.Failure
				}
			}
		}
	}

	// Generate and download the components
	err = services.UpdateComponents(c, env, repo, updateResourcesSince, t.Reset)
	if err != nil {
		c.Error(err.Error())
		if !errors.Is(err, services.UserError) {
			return inter.Failure
		}
	}

	c.Line("")
	for _, host := range env.GetExplicitHosts() {
		method := "https://"
		if env.Local {
			method = "http://"
		}
		c.Info("Website: %s%s", method, host)
		c.Info("Admin: %s%s%s\n", method, host, "/admin")
	}
	c.Line("")

	// Send event to the event bus
	event_bus.SendMessage(event_bus.Message{Type: "remote_file_processed", Message: "File processed"})

	// Scan and watch next changes
	scanner.Scanner{
		RemoteCommit: remoteCommit,
		Writer:       c.Writer(),
	}.Watch(c, env, repo)

	// The watch is preventing the code from ever getting here
	return inter.Success
}
