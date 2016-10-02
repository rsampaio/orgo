package main

import (
	"net/http"

	"golang.org/x/oauth2"
)

type GoogleHandler struct {
	config *oauth2.Config
}

func NewGoogleHandler() *GoogleHandler {
	return &GoogleHandler{}
}

func (g *GoogleHandler) HandleGoogleOauthCallback(w http.ResponseWriter, r *http.Request) {

}
