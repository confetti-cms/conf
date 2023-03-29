package scanner

import (
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"path/filepath"
	"src/app/services"
	"strings"
)

type Scanner struct {
	Verbose      bool
	RemoteCommit string
	Root         string
}

func (w Scanner) Watch() {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	// Start listening for events.
	go w.startListening(watcher)
	// Add all directories to the watcher
	w.addRecursive(watcher, w.Root)
	// Block main goroutine forever.
	<-make(chan struct{})
}

func (w Scanner) addRecursive(watcher *fsnotify.Watcher, dir string) {
	err := filepath.WalkDir(dir, func(walkPath string, d os.DirEntry, err error) error {
		if strings.Contains(walkPath, "/.") {
			return nil
		}
		if !d.IsDir() {
			return err
		}
		if w.Verbose {
			println("Watch directory: " + walkPath)
		}
		if err != nil {
			return err
		}
		err = watcher.Add(walkPath)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (w Scanner) startListening(watcher *fsnotify.Watcher) {
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
				continue
			}
			// Trim local file path
			filePath := strings.ReplaceAll(event.Name, w.Root+"/", "")
			if w.Verbose {
				log.Println("Modified file: ", event.Name, " Op:", event.Op)
			}
			if event.Op == fsnotify.Rename || event.Op == fsnotify.Remove {
				err := services.SendDeleteSource(filePath)
				if err != nil {
					println("Err SendDeleteSource:")
					println(err.Error())
				}
				continue
			}
			patch := services.GetPatchSinceCommit(w.RemoteCommit, w.Root, filePath, w.Verbose)
            services.SendPatch(filePath, patch, w.Verbose)
            // Get and save hidden files in .confetti
            services.UpsertHiddenComponentE(w.Root, filePath, w.Verbose)
            err := services.UpsertHiddenMap(w.Root, w.Verbose)
            if err != nil {
                println("Err UpsertHiddenMap:")
                println(err.Error())
            }
			success(w.Verbose)
		case err, ok := <-watcher.Errors:
			if !ok {
				continue
			}
			log.Println("error: ", err)
		}
	}
}

func success(verbose bool) {
	if !verbose {
		print(".")
	}
}
