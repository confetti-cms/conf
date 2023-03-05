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
	c.Line("confetti watch")
	c.Info("Website: https://4s89fhw0.%s.nl", pathDirs[len(pathDirs)-1])
	c.Info("Admin:   https://admin.4s89fhw0.%s.nl", pathDirs[len(pathDirs)-1])

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

	changes := services.ChangedFilesSinceLastCommit(root)

	// Send patch since latest remote commit
	for _, change := range changes {
		services.SendPatchSinceCommit(remoteCommit, root, change.Path)
	}
	c.Info("Remote server in sync")

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
				if strings.HasSuffix(event.Name, "swp") || strings.HasSuffix(event.Name, "~") {
					continue
				}
				if event.Op == fsnotify.Chmod {
					if t.Verbose {
						log.Println("Ignore "+event.Op.String()+" by file:", event.Name)
					}
					continue
				}
				// Trim local file path
				filePath := strings.ReplaceAll(event.Name, root+"/", "")
				if t.Verbose {
					log.Println("Modified file: ", event.Name, " Op:", event.Op)
                }
                if event.Op == fsnotify.Rename || event.Op == fsnotify.Remove {
                    err = services.SendDeleteSource(filePath)
                    if err != nil {
                        println("Err:")
                        println(err.Error())
                    }
                    continue
                }
				services.SendPatchSinceCommit(remoteCommit, root, filePath)
			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				log.Println("error: ", err)
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
