package commands

import (
	"log"
	"path"
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
	dir := t.Directory
	c.Info("Read directory: %s", dir)

	remoteCommit := services.GitRemoteCommit(dir)

	changes := services.ChangedFilesSinceLastCommit(dir)

	for _, change := range changes {
		patch, err := services.GetPatchSinceCommit(remoteCommit, path.Join(dir, change.Path))
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
	}

	return inter.Success
}
