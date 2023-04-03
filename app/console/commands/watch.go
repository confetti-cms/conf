package commands

import (
	"log"
	"src/app/services"
	"src/app/services/scanner"
	"strings"

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
	// Guess the domain name
	pathDirs := strings.Split(root, "/")
	c.Info("Confetti watch")
	// Get commit of the remote repository
	remoteCommit := services.GitRemoteCommit(root)
	c.Line("Sync...")
	if t.Reset {
		c.Info("Reset all components")
	}
	err := services.SendCheckout(services.CheckoutBody{
		Commit: remoteCommit,
		Reset:  t.Reset,
	})
	if err != nil {
		log.Fatal(err)
	}
	// First get the standard hidden files
	// we override the files when there are local changes
	err = services.SaveStandardHiddenFiles(root, t.Verbose)
	if err != nil {
		log.Fatal(err)
	}
	services.PatchDir(root, remoteCommit, c.Writer(), t.Verbose)
	err = services.UpsertHiddenMap(root, t.Verbose)
	if err != nil {
		log.Fatal(err)
	}
	c.Line("")
	c.Info("Website: https://4s89fhw0.%s.nl", pathDirs[len(pathDirs)-1])
	c.Info("Admin:   https://admin.4s89fhw0.%s.nl", pathDirs[len(pathDirs)-1])
	// Scan and watch next changes
	scanner.Scanner{
		Verbose:      t.Verbose,
		RemoteCommit: remoteCommit,
		Root:         root,
		Writer:       c.Writer(),
	}.Watch("")
	// The watch is preventing the code from ever getting here
	return inter.Success
}
