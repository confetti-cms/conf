package commands

import (
	"io"
	"log"
	"src/app/services"
	"src/app/services/scanner"
	"strings"
    "sync"

    "github.com/confetti-framework/framework/inter"
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

// WaitGroup is used to wait for the program to finish goroutines.
var wg sync.WaitGroup

func (t Watch) Handle(c inter.Cli) inter.ExitCode {
	root := t.Directory
	if t.Verbose {
		c.Info("Read directory: %s", root)
	}
	// Guess the domain name
	pathDirs := strings.Split(root, "/")
	c.Info("Confetti watch")
    // Get commit of the remote repository
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
	// Get patches since latest remote commits
	changes := services.ChangedFilesSinceLastCommit(root)
	// Send patches since latest remote commits
	bar := t.getBar(len(changes)*2, "Sync local changes with Confetti", c)
    wg.Add(len(changes))
	for _, change := range changes {
		change := change
		go func() {
            defer wg.Done()
			patch := services.GetPatchSinceCommit(remoteCommit, root, change.Path, t.Verbose)
			_ = bar.Add(1)
			services.SendPatch(change.Path, patch, t.Verbose)
			_ = bar.Add(1)
		}()
	}
    // Wait for the goroutines to finish.
    wg.Wait()
    c.Info("Website: https://4s89fhw0.%s.nl", pathDirs[len(pathDirs)-1])
    c.Info("Admin:   https://admin.4s89fhw0.%s.nl", pathDirs[len(pathDirs)-1])
	// Scan and watch next changes
	scanner.Scanner{
		Verbose:      t.Verbose,
		RemoteCommit: remoteCommit,
		Root:         root,
	}.Watch()
	// The watch is preventing the code from ever getting here
	return inter.Success
}

func (t Watch) getBar(total int, description string, c inter.Cli) *progressbar.ProgressBar {
	if total == 0 {
		return nil
	}
	writer := c.Writer()
	if t.Verbose {
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
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: "-",
			BarStart:      "|",
			BarEnd:        "|",
		}))
}
