package commands

import (
	"log"
	"path"
	"src/app/services"
	"strings"

	"github.com/confetti-framework/framework/inter"
	"github.com/fsnotify/fsnotify"
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

	// First send patch since latest remote commit
	for _, change := range changes {
		services.SendPatchSinceCommit(remoteCommit, root, change.Path)
	}

	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					log.Println("Not ok")
					continue
				}
				log.Println(event.Name)
				if strings.HasSuffix(event.Name, "swp") || strings.HasSuffix(event.Name, "~") {
					continue
				}
				if event.Op == fsnotify.Chmod || event.Op == fsnotify.Rename {
					log.Println("Ignore "+event.Op.String()+" by file:", event.Name)
					continue
				}
				log.Println("modified file:", event.Name)
				path := strings.ReplaceAll(event.Name, root+"/", "")
				log.Println("root:", root)
				log.Println("after trim left:", path)
				services.SendPatchSinceCommit(remoteCommit, root, path)
			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				log.Println("error:", err)
			}
		}
	}()

	// Watch the rood of the project.
	// err = watcher.Add(root)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// Watch the view directory.
	err = watcher.Add(path.Join(root, "views"))
	if err != nil {
		log.Fatal(err)
	}

	// Block main goroutine forever.
	<-make(chan struct{})

	return inter.Success
}
