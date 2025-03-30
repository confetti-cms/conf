package services

import (
	"fmt"
	"os"
	"src/config"

	"github.com/confetti-framework/framework/inter"
)

func ComposerInstall(cli inter.Cli, env Environment) error {
	// If config.Path.Root has composer.json but no vendor directory, install composer
	_, err := os.Stat(config.Path.Root + "/composer.json")
	if os.IsNotExist(err) {
		if config.App.VeryVerbose {
			cli.Info("No composer.json found in %s, skipping composer install", config.Path.Root)
		}
		return nil
	}
	_, err = os.Stat(config.Path.Root + "/vendor")
	if !os.IsNotExist(err) {
		if config.App.VeryVeryVerbose {
			cli.Info("Vendor directory found in %s, skipping composer install", config.Path.Root)
		}
		return nil
	}

	cli.Info("Composer install")

	cmd := fmt.Sprintf("cd %s && composer install --ignore-platform-reqs --no-interaction --no-progress --no-plugins", config.Path.Root)

	// Ignore the result because composer's warnings go to stderr.
	_ = StreamCommand(cmd)

	// Check if the vendor directory is created.
	// Ignore the error if the vendor directory was successfully created,
	// as PHP may return many warnings to stderr, such as "Cannot load Xdebug - it was already loaded".
	_, err = os.Stat(config.Path.Root + "/vendor")
	if os.IsNotExist(err) {
		if config.App.VeryVerbose {
			println("Vendor directory not found in %s, downloading vendor directory from remote server", config.Path.Root)
		}
		err := DownloadZip(cli, env.GetServiceUrl("confetti-cms/parser")+"/vendor", config.Path.Root+"vendor", env)
		if err != nil {
			return err
		}
	}

	return nil
}
