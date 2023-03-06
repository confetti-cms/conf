package services

import (
	"net/http"
)

type PatchBody struct {
	Path      string `json:"path"`
	Patch     string `json:"patch"`
	Untracked bool   `json:"is_untracked"`
}

func SendPatch(patch, path string, verbose bool) {
    err := SendPatchE(patch, path, verbose)
	if err != nil {
		println("Err SendPatchE:")
		println(err.Error())
		return
	}
	if verbose {
		println("Patch send: " + path)
	}
}

func SendPatchE(patch, path string, verbose bool) error {
	if patch == "" && verbose {
		println("Ignore (no change in patch): " + path)
		return nil
	}
	body := PatchBody{
		Path:  path,
		Patch: patch,
	}
	return Send("http://api.localhost/parser/source", body, http.MethodPatch)
}
