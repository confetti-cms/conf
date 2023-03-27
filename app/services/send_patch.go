package services

import (
	"net/http"
)

type PatchBody struct {
	Path      string `json:"path"`
	Patch     string `json:"patch"`
	Untracked bool   `json:"is_untracked"`
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
