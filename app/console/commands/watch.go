package commands

import (
	"log"
	"src/app/services"

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
	root := t.Directory
	c.Info("Read directory: %s", root)

	remoteCommit := services.GitRemoteCommit(root)

	changes := services.ChangedFilesSinceLastCommit(root)

	for _, change := range changes {
		patch, err := services.GetPatchSinceCommit(remoteCommit, root, change.Path)
		if err != nil {
			log.Fatal(err)
		}
		err = services.SendPatch(services.PatchBody{
			Path:  change.Path,
			Patch: patch,
		})
		if err != nil {
			log.Fatal(err)
		}
		println("Has send: " + change.Path)
	}

	return inter.Success
}
