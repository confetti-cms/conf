package services

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"src/config"
	"sync"

	"github.com/confetti-framework/framework/inter"
	"github.com/schollz/progressbar/v3"
)

type PatchBody struct {
	Path  string `json:"path"`
	Patch string `json:"patch"`
}

// WaitGroup is used to wait for the program to finish goroutines.
var wg sync.WaitGroup

const maxChanges = 300

func PatchDir(cli inter.Cli, env Environment, remoteCommit string, writer io.Writer, repo string) []string {
	// Get patches since latest remote commits
	changes := ChangedFilesSinceRemoteCommit(remoteCommit)
	changes = IgnoreHidden(changes)
	changesFiles := []string{}
	// Do not allow too many changes
	if len(changes) > maxChanges {
		cli.Error("Too many changes. Use .gitignore to ignore libraries or commit and push your changes before running this command.")
		os.Exit(1)
	}
	// Send patches since latest remote commits
	bar := getBar(len(changes)*3, "Sync local changes with Confetti", writer)
	wg.Add(len(changes))
	for _, change := range changes {
		change := change
		changesFiles = append(changesFiles, change.Path)
		go func() {
			defer wg.Done()
			_ = bar.Add(1)
			removed := RemoveIfDeleted(cli, env, change, repo)
			if removed {
				if config.App.Debug {
					println("File removed: " + change.Path)
				}
				_ = bar.Add(2)
				return
			}
			if config.App.Debug {
				println("Patch file: " + change.Path)
			}
			patch, err := GetPatchSinceCommit(remoteCommit, change.Path, change.Status == GitStatusAdded)
			if err != nil {
				_ = bar.Add(2)
				if err != ErrNewFileEmptyPatch {
					println("Err: get patch when patch dir: " + err.Error())
				}
				return
			}
			_ = bar.Add(1)
			SendPatch(cli, env, change.Path, patch, repo)
			_ = bar.Add(1)
		}()
	}
	// Wait for the goroutines to finish.
	wg.Wait()
	return changesFiles
}

func SendPatch(cli inter.Cli, env Environment, path, patch string, repo string) {
	err := SendPatchE(cli, env, path, patch, repo)
	if err != nil {
		cli.Error(err.Error())
		if !errors.Is(err, UserError) {
			PlayErrorSound()
			log.Fatal(err)
		}
		return
	}
	if config.App.Debug {
		println("Patch send: " + path)
	}
}

func SendPatchE(cli inter.Cli, env Environment, path, patch string, repo string) error {
	body := PatchBody{
		Path:  path,
		Patch: patch,
	}
	if config.App.Debug {
		println("Patch sending:", path)
	}
	url := env.GetServiceUrl("confetti-cms/parser")
	_, err := Send(cli, url+"/source", body, http.MethodPatch, env, repo)
	return err
}

func getBar(total int, description string, writer io.Writer) *progressbar.ProgressBar {
	if total == 0 {
		return nil
	}
	if config.App.Debug {
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
