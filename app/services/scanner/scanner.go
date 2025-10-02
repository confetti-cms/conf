package scanner

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"src/app/services"
	"src/app/services/event_bus"
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

	// Add root directory to the watcher
	err = watcher.Add(dir)
	if err != nil {
		log.Fatalf("Failed to add root directory %s to watcher: %v", dir, err)
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
			if config.App.VeryVeryVerbose {
				println("Ignore by Gitignore directory: " + walkPath)
			}
			continue
		}

		// Log the directory being watched if verbose mode is enabled
		if config.App.VeryVeryVerbose {
			println("Watch directory: " + walkPath)
		}

		// Add the directory to the watcher
		err := watcher.Add(walkPath)
		if err != nil {
			log.Fatalf("Failed to add directory %s to watcher: %v", walkPath, err)
		}

		// Recursive call for subdirectories
		if config.App.VeryVeryVerbose {
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
				log.Println("watcher.Events channel closed")
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
			fileInfo, err := os.Stat(event.Name)
			if err != nil {
				// If the file is added and then removed, we can get 'no such file or directory' error.
				if config.App.VeryVerbose && !os.IsNotExist(err) {
					println("Err: when check file for dir: " + err.Error())
					println(event.Name)
				}
				continue
			}
			// Not removing
			if fileInfo.IsDir() {
				if services.GitIgnored(event.Name) {
					continue
				}
				if config.App.Verbose {
					println("Patch and watch new dir: " + event.Name)
				}
				services.PatchDir(cli, env, w.RemoteCommit, w.Writer, repo)
				// Remove loading bar
				clearLines()
				clearLines()

				w.addRecursive(watcher, event.Name)
				continue
			}

			event_bus.SendMessage(event_bus.Message{Type: "local_file_changed", Message: "Local file changed"})

			patch, err := services.GetPatchSinceCommitE(w.RemoteCommit, file, eventIs(event, fsnotify.Create))
			if err != nil {
				if err != services.ErrNewFileEmptyPatch {
					println("Err: get patch when scanner start listening: " + err.Error())
				}
				// Send event to the event bus
				event_bus.SendMessage(event_bus.Message{Type: "error", Message: fmt.Sprintf("Error getting patch for file %s: See your terminal for more information", file)})
				continue
			}

			if patch == "" && config.App.Verbose {
				fmt.Printf("Warning: patch is empty in startListening, file: %s, this is fine if the user undo all changes in a file\n", file)
			}

			services.SendPatch(cli, env, file, patch, repo, 30*time.Second)
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

			// Send event to the event bus
			event_bus.SendMessage(event_bus.Message{Type: "remote_file_processed", Message: "File processed"})

			clearLines()
			fmt.Printf("Latest sync: %s\n", time.Now().Format("2006-01-02 15:04:05"))
			fmt.Printf("\033[1;34m%s\033[0m", file)
			if config.App.Verbose {
				fmt.Printf("\n\n")
			}
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

func clearLines() {
	// Clear the file line:
	// \033[2K clears the current line.
	// \r returns the cursor to the start of the line.
	fmt.Printf("\033[2K\r")
	// Move the cursor up one line and clear that line (the sync line)
	fmt.Printf("\033[A\033[2K\r")
}
