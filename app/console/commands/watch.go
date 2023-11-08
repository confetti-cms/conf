package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"src/app/services"
	"src/app/services/scanner"
	"src/config"

	"github.com/confetti-framework/errors"

	"github.com/confetti-framework/framework/inter"
)

type Watch struct {
	Directory   string `short:"p" flag:"path" description:"Root directory of the Git repository"`
	Verbose     bool   `short:"v" flag:"verbose" description:"Show events"`
	Reset       bool   `short:"r" flag:"reset" description:"All files are parsed again"`
	Environment string `short:"e" flag:"env" description:"The environment key in the app_config.json5 file, default 'dev'"`
}

func (t Watch) Name() string {
	return "watch"
}

func (t Watch) Description() string {
	return "Keeps your local files in sync with the server."
}

func (t Watch) Handle(c inter.Cli) inter.ExitCode {
	config.App.Debug = t.Verbose
	root, err := t.getDirectoryOrCurrent()
	if err != nil {
		c.Error(err.Error())
		return inter.Failure
	}
	if config.App.Debug {
		c.Info("Use directory: %s", root)
	}
	c.Info("Confetti watch")
	env, err := services.GetEnvironmentByInput(c, root, t.Environment)
	if err != nil {
		c.Error(err.Error())
		return inter.Failure
	}
	// Get commit of the remote repository
	remoteCommit := services.GitRemoteCommit(root)
	if t.Reset {
		c.Info("Reset all components")
	}
	repo, err := services.GetRepositoryName(root)
	if err != nil {
	    c.Error(err.Error())
	    return inter.Failure
	}
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
	services.PatchDir(c, env, root, remoteCommit, c.Writer(), repo)
	// Get the standard hidden files
	err = services.FetchHiddenFiles(c, env, root, repo)
	if err != nil {
		c.Error(err.Error())
		if !errors.Is(err, services.UserError) {
			return inter.Failure
		}
	}
	// Remove loading bar
	fmt.Printf("\r                                                                      ")
	c.Line("")
	for _, host := range env.GetExplicitHosts() {
		c.Info("Website: http://%s", host)
		c.Info("Admin: http://%s%s\n", host, "/admin")
	}
	// Scan and watch next changes
	scanner.Scanner{
		RemoteCommit: remoteCommit,
		Root:         root,
		Writer:       c.Writer(),
	}.Watch(c, env, "", repo)
	// The watch is preventing the code from ever getting here
	return inter.Success
}

func (t Watch) getDirectoryOrCurrent() (string, error) {
	if t.Directory != "" {
		if _, err := os.Stat(filepath.Join(t.Directory, ".git")); os.IsNotExist(err) {
			return "", errors.New("The specified directory is incorrect. Please ensure that the given directory is correct.")
		}
		return t.Directory, nil
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
