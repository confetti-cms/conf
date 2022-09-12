package commands

import (
	"github.com/confetti-framework/framework/inter"
)

type Watch struct {
	Directory string `short:"d" flag:"directory" description:"Root directory of the Git repository" required:"true"`
}

func (t Watch) Name() string {
	return "watch"
}

func (t Watch) Description() string {
	return "Keeps your local files in sync with the server."
}

func (t Watch) Handle(c inter.Cli) inter.ExitCode {
	
	c.Info("Read directory: %s", t.Directory)
	return inter.Success
}
