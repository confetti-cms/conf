package services

import (
	"github.com/schollz/progressbar/v3"
	"io"
	"net/http"
	"sync"
)

type PatchBody struct {
	Path      string `json:"path"`
	Patch     string `json:"patch"`
	Untracked bool   `json:"is_untracked"`
}

// WaitGroup is used to wait for the program to finish goroutines.
var wg sync.WaitGroup

func PatchDir(root string, remoteCommit string, writer io.Writer, verbose bool) {
	// Get patches since latest remote commits
	changes := ChangedFilesSinceLastCommit(root)
	changes = IgnoreHidden(changes)
	// Send patches since latest remote commits
	bar := getBar(len(changes)*2, "Sync local changes with Confetti", writer, verbose)
	wg.Add(len(changes))
	for _, change := range changes {
		change := change
		go func() {
			defer wg.Done()
			removed := RemoveIfDeleted(change, root)
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
			patch := GetPatchSinceCommit(remoteCommit, root, change.Path, verbose)
			_ = bar.Add(1)
			SendPatch(change.Path, patch, verbose)
			// Get and save hidden files in .confetti
			UpsertHiddenComponentE(root, change.Path, verbose)
			_ = bar.Add(1)
		}()
	}
	// Wait for the goroutines to finish.
	wg.Wait()
}

func SendPatch(path, patch string, verbose bool) {
	err := SendPatchE(path, patch, verbose)
	if err != nil {
		println("Err SendPatchE:")
		println(err.Error())
		return
	}
}

func SendPatchE(path, patch string, verbose bool) error {
	if patch == "" {
		if verbose {
			println("Ignore (no change in patch): " + path)
		}
		return nil
	}
	body := PatchBody{
		Path:  path,
		Patch: patch,
	}
	if verbose {
		println("Patch sending:", path)
	}
	_, err := Send("http://api.localhost/parser/source", body, http.MethodPatch)
    if verbose {
        println("Patch send: " + path)
    }
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
