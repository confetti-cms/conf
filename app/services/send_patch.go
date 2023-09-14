package services

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/confetti-framework/framework/inter"
	"github.com/schollz/progressbar/v3"
)

type PatchBody struct {
	Path      string `json:"path"`
	Patch     string `json:"patch"`
	Untracked bool   `json:"is_untracked"`
}

// WaitGroup is used to wait for the program to finish goroutines.
var wg sync.WaitGroup

const maxChanges = 100

func PatchDir(cli inter.Cli, env Environment, root string, remoteCommit string, writer io.Writer, verbose bool) {
	// Get patches since latest remote commits
	changes := ChangedFilesSinceRemoteCommit(root, remoteCommit)
	changes = IgnoreHidden(changes)
	// Do not allow too many changes
	if len(changes) > maxChanges {
		_, err := writer.Write([]byte("Too many changes. Use .gitignore to ignore libraries or commit and push your changes before running this command."))
		if err != nil {
			panic(err)
		}
		os.Exit(1)
	}
	// Send patches since latest remote commits
	bar := getBar(len(changes)*2, "Sync local changes with Confetti", writer, verbose)
	wg.Add(len(changes))
	for _, change := range changes {
		change := change
		go func() {
			defer wg.Done()
			removed := RemoveIfDeleted(cli, env, change, root)
			if removed {
				if verbose {
					println("File removed: " + change.Path)
				}
				_ = bar.Add(2)
				return
			}
			if verbose {
				println("Patch file: " + change.Path)
			}
			patch := GetPatchSinceCommit(remoteCommit, root, change.Path, change.Status == GitStatusAdded, verbose)
			_ = bar.Add(1)
			SendPatch(cli, env, change.Path, patch, verbose)
			_ = bar.Add(1)
		}()
	}
	// Wait for the goroutines to finish.
	wg.Wait()
}

func SendPatch(cli inter.Cli, env Environment, path, patch string, verbose bool) {
	err := SendPatchE(cli, env, path, patch, verbose)
	if err != nil {
		cli.Error(err.Error())
		if !errors.Is(err, UserError) {
			log.Fatal(err)
		}
		return
	}
	if verbose {
		println("Patch send: " + path)
	}
}

func SendPatchE(cli inter.Cli, env Environment, path, patch string, verbose bool) error {
	body := PatchBody{
		Path:  path,
		Patch: patch,
	}
	if verbose {
		println("Patch sending:", path)
	}
	url := env.GetServiceUrl("confetti-cms/parser")
	_, err := Send(cli, url+"/source", body, http.MethodPatch)
	return err
}

func getBar(total int, description string, writer io.Writer, verbose bool) *progressbar.ProgressBar {
	if total == 0 {
		return nil
	}
	if verbose {
		// Ignore progressbar in verbose mode
		writer = io.Discard
	}
	return progressbar.NewOptions(total,
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
		}))
}
