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

	// We can't/don't want to use radovskyb/watcher. It is unreliable and gives event CREATED for files that are not created. We use fsnotify/fsnotify instead.
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
	// Do not use filepath.WalkDir. WalkDir will give all the directories and files in the directory.
	// When the project has vendor directories, it is not efficient to check all the directories if they has to be watched.
	// Better we go through the directory tree ourselves. That seems mutch more efficient.
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("Failed to read directory %s: %v", dir, err)
	}

	for _, entry := range entries {
		walkPath := filepath.Join(dir, entry.Name())

		// Ignore files based on services.IgnoreFile
		if services.IgnoreFile(walkPath) {
			continue
		}

		// Skip if entry is not a directory
		if !entry.IsDir() {
			continue
		}

		// Skip Git-ignored directories
		if services.GitIgnored(walkPath) {
			println("Ignore by Gitignore directory: " + walkPath)
			continue
		}

		// Log the directory being watched if verbose mode is enabled
		if config.App.Verbose {
			println("Watch directory: " + walkPath)
		}

		// Add the directory to the watcher
		err := watcher.Add(walkPath)
		if err != nil {
			log.Fatalf("Failed to add directory %s to watcher: %v", walkPath, err)
		}

		// Recursive call for subdirectories
		if config.App.Verbose {
			println("Recursive add: " + walkPath)
		}
		w.addRecursive(watcher, walkPath)
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
			if config.App.VeryVerbose {
				log.Println("Modified file: ", event.Name, " Op:", event.Op)
			}
			// Removed (by removing or renaming)
			if eventIs(event, fsnotify.Rename) || eventIs(event, fsnotify.Remove) {
				if config.App.VeryVerbose {
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
				if config.App.Verbose {
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
				if config.App.Verbose {
					println("Patch is empty in startListening !!! file: " + file)
				}
				continue
			}

			services.SendPatch(cli, env, file, patch, repo)
			if services.IsBaseComponent(file) {
				if config.App.VeryVerbose {
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
			if config.App.Verbose {
				ln = "\n"
			}
			fmt.Printf("\rLatest sync: %s \033[1;34m%s\033[0m                              %s", time.Now().Format("2006-01-02 15:04:05"), file, ln)
		case err, ok := <-watcher.Errors:
			if !ok {
				continue
			}
			log.Println("error watching:", err)
		}
	}
}

func eventIs(given fsnotify.Event, expect fsnotify.Op) bool {
	return strings.Contains(given.Op.String(), expect.String()[1:])
}
