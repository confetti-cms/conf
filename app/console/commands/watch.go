package commands

import (
	"log"
	"src/app/services"
	"src/app/services/scanner"
	"src/config"

	"github.com/confetti-framework/framework/inter"
)

type Watch struct {
	Directory string `short:"p" flag:"path" description:"Root directory of the Git repository" required:"true"`
	Verbose   bool   `short:"v" flag:"verbose" description:"Show events"`
	Reset     bool   `short:"r" flag:"reset" description:"All files are parsed again"`
}

func (t Watch) Name() string {
	return "watch"
}

func (t Watch) Description() string {
	return "Keeps your local files in sync with the server."
}

func (t Watch) Handle(c inter.Cli) inter.ExitCode {
	root := t.Directory
	if t.Verbose {
		c.Info("Read directory: %s", root)
	}
	c.Info("Confetti watch")
	// Get commit of the remote repository
	remoteCommit := services.GitRemoteCommit(root)
	if t.Reset {
		c.Info("Reset all components")
	}
	err := services.SendCheckout(
		c,
		services.CheckoutBody{
			Commit: remoteCommit,
			Reset:  t.Reset,
		})
	if err != nil {
		log.Fatal(err)
	}
	services.PatchDir(c, root, remoteCommit, c.Writer(), t.Verbose)
	// Get the standard hidden files
	err = services.FetchHiddenFiles(c, root, t.Verbose)
	if err != nil {
		log.Fatal(err)
	}
	c.Line("")
	c.Info("Website: http://%s", config.App.Host)
	c.Info("Admin:   http://%s", config.App.Host+"/admin")
	// Scan and watch next changes
	scanner.Scanner{
		Verbose:      t.Verbose,
		RemoteCommit: remoteCommit,
		Root:         root,
		Writer:       c.Writer(),
	}.Watch(c, "")
	// The watch is preventing the code from ever getting here
	return inter.Success
}
