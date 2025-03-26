package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"src/app/services"
	"src/app/services/event_bus"
	"src/app/services/scanner"
	"src/config"
	"strings"
	"time"

	"github.com/confetti-framework/errors"

	"github.com/confetti-framework/framework/inter"
)

type Watch struct {
	Directory       string `short:"p" flag:"path" description:"Root directory of the Git repository"`
	Verbose         bool   `short:"v" description:"Show events"`
	VeryVerbose     bool   `short:"vv" description:"Show more events"`
	VeryVeryVerbose bool   `short:"vvv" description:"Show all events"`
	Reset           bool   `short:"r" flag:"reset" description:"All files are parsed again"`
	Environment     string `short:"n" flag:"name" description:"The environment name in the app_config.json5 file, default 'dev'"`
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
	root, err := t.getDirectoryOrCurrent()
	if err != nil {
		c.Error(err.Error())
		return inter.Failure
	}
	config.Path.Root = root
	if config.App.Verbose {
		c.Info("Use directory: %s", root)
	}
	c.Info("Confetti watch")
	env, err := services.GetEnvironmentByInput(c, t.Environment)
	if err != nil {
		c.Error(fmt.Sprintf("Error getting environment: %s", err))
		return inter.Failure
	}

	// Open the event bus server
	go event_bus.Publish()

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
	}, repo)
	if err != nil {
		c.Error(err.Error())
		if !errors.Is(err, services.UserError) {
			return inter.Failure
		}
	}

	// Apply all local changes
	filesToSync := services.PatchDir(c, env, remoteCommit, c.Writer(), repo)
	// Remove loading bar
	fmt.Printf("\r                                                                      \n")

	// Run composer install locally if vendor directory is missing
	err = services.ComposerInstall(c, env)
	if err != nil {
		c.Error(err.Error())
		if !errors.Is(err, services.UserError) {
			return inter.Failure
		}
	}

	// Parse all base components (other components wil extend this components)
	err = services.ParseBaseComponents(c, env, repo)
	if err != nil {
		c.Error(err.Error())
		if !errors.Is(err, services.UserError) {
			return inter.Failure
		}
	}

	// Parse components
	for _, file := range filesToSync {
		err = services.ParseComponent(c, env, services.ParseComponentBody{File: file}, repo)
		if err != nil {
			c.Error(err.Error())
			if !errors.Is(err, services.UserError) {
				return inter.Failure
			}
		}
	}

	// Generate and download the components
	err = services.UpdateComponents(c, env, repo, updateResourcesSince)
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

func (t Watch) getDirectoryOrCurrent() (string, error) {
	if t.Directory != "" {
		if _, err := os.Stat(filepath.Join(t.Directory, ".git")); os.IsNotExist(err) {
			return "", errors.New("The specified directory is incorrect. Please ensure that the given directory is correct.")
		}
		return strings.TrimRight(t.Directory, "/") + "/", nil
	}
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(path, ".git")); os.IsNotExist(err) {
		return "", errors.New("You are not running this command in the correct location. Please ensure that you are running the command in the correct Git repository.")
	}
	return path, nil
}
