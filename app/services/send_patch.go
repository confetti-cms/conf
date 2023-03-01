package services

import (
    "net/http"
)

type PatchBody struct {
	Path      string `json:"path"`
	Patch     string `json:"patch"`
	Untracked bool   `json:"is_untracked"`
}

func SendPatch(requestBody PatchBody) error {
    return Send("http://api.localhost/parser/source", requestBody, http.MethodPatch)
}
