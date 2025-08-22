package services

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"src/config"
	"sync"
	"time"

	"github.com/confetti-framework/framework/inter"
	"github.com/schollz/progressbar/v3"
)

type PatchBody struct {
	Path  string `json:"path"`
	Patch string `json:"patch"`
}

// WaitGroup is used to wait for the program to finish goroutines.
var wg sync.WaitGroup

const maxChanges = 500

func ClearLines() {
	// Clear the file line:
	// \033[2K clears the current line.
	// \r returns the cursor to the start of the line.
	fmt.Printf("\033[2K\r")
	// Move the cursor up one line and clear that line (the sync line)
	fmt.Printf("\033[A\033[2K\r")
}

func PatchDir(cli inter.Cli, env Environment, remoteCommit string, writer io.Writer, repo string) []string {
	// Get patches since latest remote commits
	changes := ChangedFilesSinceRemoteCommit(remoteCommit)
	changes = IgnoreHidden(changes)
	changesFiles := []string{}
	// Do not allow too many changes
	if len(changes) > maxChanges {
		cli.Error("Too many changes (" + fmt.Sprintf("%d", len(changes)) + "). Use .gitignore to ignore libraries or commit and push your changes before running this command.")
		PlayErrorSound()
		os.Exit(1)
	}
	// Send patches since latest remote commits
	ClearLines()
	// Message if more than 50 changes, commit and push your changes an rerun this command
	message := "Sync local changes"
	if len(changes) > 50 {
		message = fmt.Sprintf("Found %d changes. For faster results, commit & push, then rerun.", len(changes))
	}
	fmt.Println(message)

	bar := getBar(len(changes), "", writer)
	wg.Add(len(changes))
	for _, change := range changes {
		change := change
		changesFiles = append(changesFiles, change.Path)
		go func() {
			defer wg.Done()
			if config.App.VeryVerbose {
				println("Patch sending: " + change.Path)
			}
			removed := RemoveIfDeleted(cli, env, change, repo)
			if removed {
				if config.App.VeryVerbose {
					println("File removed: " + change.Path)
				}
				_ = bar.Add(1)
				return
			}
			if config.App.VeryVerbose {
				println("Patch file: " + change.Path)
			}
			patch, err := GetPatchSinceCommit(remoteCommit, change.Path, change.Status == GitStatusAdded)
			if err != nil {
				if err != ErrNewFileEmptyPatch {
					println("Err: get patch when patch dir: " + err.Error())
				}
				_ = bar.Add(1)
				return
			}
			if patch == "" && config.App.VeryVerbose {
				fmt.Printf("Warning: patch is empty in PatchDir, file: %s, this is fine if the user undo all changes in a file\n", change.Path)
			}
			SendPatch(cli, env, change.Path, patch, repo, time.Duration(5+len(changes))*time.Second)
			_ = bar.Add(1)
		}()
	}
	// Wait for the goroutines to finish.
	wg.Wait()
	return changesFiles
}

func SendPatch(cli inter.Cli, env Environment, path, patch string, repo string, timeout time.Duration) {
	err := SendPatchE(cli, env, path, patch, repo, timeout)
	if err != nil {
		cli.Error(err.Error())
		if !errors.Is(err, UserError) {
			PlayErrorSound()
			fmt.Printf("Error sending patch for %s: %s\n", path, err.Error())
		}
		return
	}
	if config.App.VeryVerbose {
		println("Patch sent:", path)
	}
}

func SendPatchE(cli inter.Cli, env Environment, path, patch string, repo string, timeout time.Duration) error {
	body := PatchBody{
		Path:  path,
		Patch: patch,
	}
	if config.App.VeryVeryVerbose {
		println("Patch sending:", path)
	}
	url := env.GetServiceUrl("confetti-cms/parser")
	_, err := Send(cli, url+"/source", body, http.MethodPatch, env, repo, timeout)
	return err
}

func getBar(total int, description string, writer io.Writer) *progressbar.ProgressBar {
	if total == 0 {
		return nil
	}
	if config.App.VeryVerbose {
		// AllTime progressbar in verbose mode
		writer = io.Discard
	}
	return progressbar.NewOptions(
		total,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetWriter(writer),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]|[reset]",
			SaucerPadding: "-",
			BarStart:      "|",
			BarEnd:        "|",
		}),
	)
}
