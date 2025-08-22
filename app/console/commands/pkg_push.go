package commands

import (
	"fmt"
	"os"
	"src/app/services"
	"src/config"

	"github.com/confetti-framework/framework/inter"
)

type PkgPush struct {
	Directory       string `short:"dir" flag:"directory" description:"Root directory of the project, defaults to the current directory"`
	Package         string `short:"p" flag:"package" description:"The package to push, e.g. 'confetti-cms/text'"`
	Verbose         bool   `short:"v" description:"Show events"`
	VeryVerbose     bool   `short:"vv" description:"Show more events"`
	VeryVeryVerbose bool   `short:"vvv" description:"Show all events"`
}

func (p PkgPush) Name() string {
	return "pkg:push"
}

func (p PkgPush) Description() string {
	return "Pushes the latest changes to the remote package repository."
}

func (p PkgPush) Handle(c inter.Cli) inter.ExitCode {
	config.App.Verbose = p.Verbose || p.VeryVerbose || p.VeryVeryVerbose
	config.App.VeryVerbose = p.VeryVerbose || p.VeryVeryVerbose
	config.App.VeryVeryVerbose = p.VeryVeryVerbose
	root, err := getDirectoryOrCurrent(p.Directory)
	if err != nil {
		c.Error(err.Error())
		return inter.Failure
	}
	config.Path.Root = root

	if config.App.Verbose {
		c.Info("Use directory: %s", root)
	}
	fmt.Println("\n\033[34mConfetti pkg:push\n\033[0m") // blue
	if p.Package == "" {
		fmt.Fprintln(os.Stderr, "Error: -pkg or --package flag is required")
		services.PlayErrorSound()
		os.Exit(1)
	}

	// Check if the package contains any files
	if pkgDirExists(p.Package) && !pkgHasFiles(p.Package) {
		c.Error(fmt.Sprintf("Error: Package directory '%s' is empty. Cannot push an empty package.", p.Package))
		return inter.Failure
	}

	// Check if the package directory exists
	if !pkgDirExists(p.Package) {
		c.Error(fmt.Sprintf("Error: Package directory '%s' does not exist.", p.Package))
		return inter.Failure
	}

	if config.App.Verbose {
		c.Line("Package directory exists: %s", p.Package)
	}

	// Push the package to the remote repository
	err = services.PushPackage(p.Package)

	// The watch is preventing the code from ever getting here
	return inter.Success
}
