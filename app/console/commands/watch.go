package commands

import (
    "github.com/spf13/cast"
    "log"
	"path"
	"src/app/services"
	"strings"

	"github.com/confetti-framework/framework/inter"
	"github.com/fsnotify/fsnotify"
	"github.com/schollz/progressbar/v3"
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

    bar := progressbar.Default(cast.ToInt64(len(changes)) * 2, "Sync local changes with Confetti")
	// Send patch since latest remote commitS
    for _, change := range changes {
        change := change
        go func() {
    		patch := services.GetPatchSinceCommit(remoteCommit, root, change.Path, t.Verbose)
            _ = bar.Add(1)
            services.SendPatch(change.Path, patch, t.Verbose)
            _ = bar.Add(1)
            if bar.IsFinished() {
                c.Info("Remote server in sync")
            }
        }()
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
						println("Err SendDeleteSource:")
						println(err.Error())
					}
					continue
				}
                patch := services.GetPatchSinceCommit(remoteCommit, root, filePath, t.Verbose)
                services.SendPatch(filePath, patch, t.Verbose)
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
