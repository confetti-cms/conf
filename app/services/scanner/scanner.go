package scanner

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"src/app/services"
	"strings"

	"github.com/confetti-framework/framework/inter"
	"github.com/fsnotify/fsnotify"
)

type Scanner struct {
	Verbose      bool
	RemoteCommit string
	Root         string
	Writer       io.Writer
}

func (w Scanner) Watch(cli inter.Cli, env services.Environment, dir string) {
	if dir == "" {
		dir = w.Root
	}
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	// Start listening for events.
	go w.startListening(cli, watcher, env)
	// Add all directories to the watcher
	w.addRecursive(watcher, dir)
	// Block main goroutine forever.
	<-make(chan struct{})
}

func (w Scanner) addRecursive(watcher *fsnotify.Watcher, dir string) {
	err := filepath.WalkDir(dir, func(walkPath string, d os.DirEntry, err error) error {
		if services.IgnoreFile(walkPath) {
			return nil
		}
		if !d.IsDir() {
			return err
		}
		if services.GitIgnored(w.Root, walkPath) {
			return nil
		}
		if w.Verbose {
			println("Watch directory: " + walkPath)
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

func (w Scanner) startListening(cli inter.Cli, watcher *fsnotify.Watcher, env services.Environment) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("Not ok")
				continue
			}
			// Ignore hidden files and directories
			if services.IgnoreFile(event.Name) {
				continue
			}
			if event.Op == fsnotify.Chmod {
				continue
			}
			// Trim local file path
			file := strings.ReplaceAll(event.Name, w.Root+"/", "")
			if w.Verbose {
				log.Println("Modified file: ", event.Name, " Op:", event.Op)
			}
			// Removed (by removing or renaming)
			if eventIs(event, fsnotify.Rename) || eventIs(event, fsnotify.Remove) {
				if w.Verbose {
					println("Send delete Source: " + file)
				}
				err := services.SendDeleteSource(cli, env, file)
				if err != nil {
					println("Err: SendDeleteSource:")
					println(err.Error())
				}
				if services.IsHiddenFileGenerator(file) {
					err = services.FetchHiddenFiles(cli, env, w.Root, w.Verbose)
					if err != nil {
						cli.Error(err.Error())
						if !errors.Is(err, services.UserError) {
							log.Fatal(err)
						}
					}
				}
				continue
			}
			// Not removing
			fileInfo, err := os.Stat(event.Name)
			if err != nil {
				println("Err: when check file for dir: " + err.Error())
				println(event.Name)
				continue
			}
			if fileInfo.IsDir() {
				if services.GitIgnored(w.Root, event.Name) {
					continue
				}
				if w.Verbose {
					println("Patch and watch new dir: " + event.Name)
				}
				services.PatchDir(cli, env, w.Root, w.RemoteCommit, w.Writer, w.Verbose)
				w.addRecursive(watcher, event.Name)
				continue
			}
			patch := services.GetPatchSinceCommit(w.RemoteCommit, w.Root, file, eventIs(event, fsnotify.Create), w.Verbose)
			services.SendPatch(cli, env, file, patch, w.Verbose)
			if services.IsHiddenFileGenerator(file) {
				err = services.FetchHiddenFiles(cli, env, w.Root, w.Verbose)
				if err != nil {
					cli.Error(err.Error())
					if !errors.Is(err, services.UserError) {
						log.Fatal(err)
					}
				}
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

func eventIs(given fsnotify.Event, expect fsnotify.Op) bool {
	return strings.Contains(given.Op.String(), expect.String()[1:])
}

func success(verbose bool) {
	if !verbose {
		print(".")
	}
}
