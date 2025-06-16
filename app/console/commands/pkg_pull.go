package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"src/app/services"
	"src/config"

	"github.com/confetti-framework/framework/inter"
)

type PkgPull struct {
	Directory       string `short:"p" flag:"path" description:"Root directory of the Git repository"`
	Environment     string `short:"n" flag:"name" description:"The environment name in the config.json5 file, default 'dev'"`
	Package         string `short:"pkg" flag:"package" description:"The package to pull, e.g. 'confetti-cms/office'"`
	Verbose         bool   `short:"v" description:"Show events"`
	VeryVerbose     bool   `short:"vv" description:"Show more events"`
	VeryVeryVerbose bool   `short:"vvv" description:"Show all events"`
}

func (p PkgPull) Name() string {
	return "pkg:pull"
}

func (p PkgPull) Description() string {
	return "Pulls the latest changes for a package."
}

func (p PkgPull) Handle(c inter.Cli) inter.ExitCode {
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
	fmt.Println("\n\033[34mConfetti pkg:pull\n\033[0m") // blue
	if p.Package == "" {
		fmt.Fprintln(os.Stderr, "Error: --package or -pkg flag is required")
		os.Exit(1)
	}

	if p.pkgDirExists() && !p.pkgHasFiles() {
		// Remove the empty package directory since it is confusing for the system
		c.Line("Removed empty package directory: %s", p.Package)
		err = services.RemovePackage(p.Package, "Removed empty package directory sdk/"+p.Package)
		if err != nil {
			c.Error(fmt.Sprintf("Error committing changes for package %s: %s", p.Package, err))
			return inter.Failure
		}
	}

	// Check if the environment has changes
	if config.App.Verbose {
		c.Line("Checking for modifications...")
	}
	changes, err := services.HasModifications()
	if err != nil {
		c.Error(err.Error())
		return inter.Failure
	} else if len(changes) > 0 {
		if len(changes) == 1 {
			c.Info("Your project has 1 modification (%s), please commit or stash it first.", changes[0])
		} else {
			c.Info("Your project has %d modifications (e.g. %s), please commit or stash them first.", len(changes), changes[0])
		}
		return inter.Failure
	}
	pkgDirIsNew := !p.pkgDirExists()

	// Check if the package directory exists
	if !p.pkgDirExists() {
		c.Line("Package directory does not exist, trying to restore if it was pulled in the past...")
		restored, err := services.RestoreDirectory(p.Package)
		if err != nil {
			c.Error(fmt.Sprintf("Error restoring directory for package `%s`: %s", p.Package, err))
			return inter.Failure
		}
		if restored {
			c.Info("Restored package %s successfully.", p.Package)
		}
	}

	// If the package directory still does not exist, add it as a new package
	if !p.pkgDirExists() {
		c.Line("Package directory does not exist, pulling it for the first time...")
		err := services.AddNewPackage(p.Package)
		if err != nil {
			c.Error(fmt.Sprintf("Error adding new package %s: %s", p.Package, err))
			return inter.Failure
		}
	}

	// Pull the latest changes for the package
	err = services.PullLatestChanges(p.Package)
	if err != nil {
		c.Error(fmt.Sprintf("Error pulling latest changes for package %s: %s", p.Package, err))
		return inter.Failure
	}

	// Check if the package directory has a composer.json file
	if config.App.Verbose {
		c.Line("Checking if package directory has a composer.json file...")
	}
	pkgDir := filepath.Join(config.Path.Root, "pkg", p.Package)
	_, err = os.Stat(filepath.Join(pkgDir, "composer.json"))
	if os.IsNotExist(err) {
		services.PrintPackageNoComposerMessage(p.Package)
		return inter.Failure
	} else if err != nil {
		c.Error(fmt.Sprintf("Error checking for composer.json in package directory %s: %s", pkgDir, err))
		return inter.Failure
	}

	// Update composer.json and add the package to the autoloader
	err = services.UpdateComposer(p.Package)
	if err != nil {
		c.Error("Failed to update composer.json and add the package to the autoloader: %s", err)

		return inter.Failure
	}

	// Add the package to the Git index
	err = services.CommitChanges(p.Package, "Added package to composer "+p.Package)
	if err != nil {
		c.Error(fmt.Sprintf("Error committing changes for package %s: %s", p.Package, err))
		return inter.Failure
	}

	// Print success message
	if pkgDirIsNew {
		services.PrintPackageInstalledMessage(p.Package)
	} else {
		services.PrintPackagePulledMessage(p.Package)
	}

	// The watch is preventing the code from ever getting here
	return inter.Success
}

func (p PkgPull) pkgDirExists() bool {
	dir := filepath.Join(config.Path.Root, "pkg", p.Package)
	if config.App.VeryVeryVerbose {
		fmt.Printf("Checking if package directory exists: %s\n", dir)
	}
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		if config.App.VeryVeryVerbose {
			fmt.Printf("Package directory does not exist: %s\n", dir)
		}
		return false
	}
	if config.App.VeryVeryVerbose {
		fmt.Printf("Package directory exists: %s\n", dir)
	}
	return info.IsDir()
}

// pkgHasContent checks if the package directory has any files or subdirectories.
func (p PkgPull) pkgHasFiles() bool {
	files, err := os.ReadDir(filepath.Join(config.Path.Root, "pkg", p.Package))
	if err != nil {
		return false
	}
	return len(files) > 0
}
