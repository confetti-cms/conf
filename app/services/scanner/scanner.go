package scanner

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"src/app/services"
	"src/config"
	"strings"
	"time"

	"github.com/confetti-framework/framework/inter"
	"github.com/fsnotify/fsnotify"
)

type Scanner struct {
	RemoteCommit string
	Writer       io.Writer
}

func (w Scanner) Watch(cli inter.Cli, env services.Environment, repo string) {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	// Start listening for events.
	go w.startListening(cli, watcher, env, repo)
	// Add all directories to the watcher
	w.addRecursive(watcher, config.Path.Root)
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
		if services.GitIgnored(walkPath) {
			return nil
		}
		if config.App.Debug {
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

func (w Scanner) startListening(cli inter.Cli, watcher *fsnotify.Watcher, env services.Environment, repo string) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("Not ok")
				continue
			}
			// AllTime hidden files and directories
			if services.IgnoreFile(event.Name) {
				continue
			}
			if event.Op == fsnotify.Chmod {
				continue
			}
			// Trim local file path
			file := strings.ReplaceAll(event.Name, config.Path.Root, "")
			if config.App.Debug {
				log.Println("Modified file: ", event.Name, " Op:", event.Op)
			}
			// Removed (by removing or renaming)
			if eventIs(event, fsnotify.Rename) || eventIs(event, fsnotify.Remove) {
				if config.App.Debug {
					println("Send delete Source: " + file)
				}
				err := services.SendDeleteSource(cli, env, file, repo)
				if err != nil {
					println("Err: SendDeleteSource:")
					println(err.Error())
				}
				if services.IsBaseComponent(file) {
					err = services.ParseBaseComponents(cli, env, repo)
					if err != nil {
						cli.Error(err.Error())
						if !errors.Is(err, services.UserError) {
							log.Fatal(err)
						}
					}
				}
				services.ResourceMayHaveChanged()
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
				if services.GitIgnored(event.Name) {
					continue
				}
				if config.App.Debug {
					println("Patch and watch new dir: " + event.Name)
				}
				services.PatchDir(cli, env, w.RemoteCommit, w.Writer, repo)
				// Remove loading bar
				fmt.Printf("\r                                                                      ")

				w.addRecursive(watcher, event.Name)
				continue
			}

			patch, err := services.GetPatchSinceCommitE(w.RemoteCommit, file, eventIs(event, fsnotify.Create))
			if err != nil {
				if err != services.ErrNewFileEmptyPatch {
					println("Err: get patch when scanner start listening: " + err.Error())
				}
				continue
			}
			if patch == "" {
				if config.App.Debug {
					println("Patch is empty !!! file: " + file)
				}
				continue
			}

			services.SendPatch(cli, env, file, patch, repo)
			if services.IsBaseComponent(file) {
				if config.App.Debug {
					println("Base component is changed")
				}
				err = services.ParseBaseComponents(cli, env, repo)
				if err != nil {
					cli.Error(err.Error())
				}
			}

			// Parse components
			err = services.ParseComponent(cli, env, services.ParseComponentBody{File: file}, repo)
			if err != nil {
				cli.Error(err.Error())
				if !errors.Is(err, services.UserError) {
					log.Fatal(err)
				}
			}

			// Set the flag that the resources may have changed
			services.ResourceMayHaveChanged()

			ln := ""
			if config.App.Debug {
				ln = "\n"
			}
			fmt.Printf("\rLatest sync: %s \033[1;34m%s\033[0m                              %s", time.Now().Format("2006-01-02 15:04:05"), file, ln)
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
