package main

import (
	"context"
	"log"
	"net/http"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
	oauth2api "google.golang.org/api/oauth2/v2"
)

type GoogleHandler struct {
	oauthConfig *oauth2.Config
}

func NewGoogleHandler(apiKey, apiSecret, redirectURL string) *GoogleHandler {
	return &GoogleHandler{
		oauthConfig: &oauth2.Config{
			ClientID:     apiKey,
			ClientSecret: apiSecret,
			RedirectURL:  "postmessage",
			Endpoint:     google.Endpoint,
			Scopes:       []string{calendar.CalendarScope, oauth2api.UserinfoEmailScope},
		},
	}
}
func (g *GoogleHandler) AuthCodeURL() string {
	return g.oauthConfig.AuthCodeURL(uuid.NewV4().String(), oauth2.AccessTypeOffline)
}

// HandleGoogleOauthCallback will receive the AuthCode from when the user authorize the app
// with the code we should store the token_id and the code for future use and validation
func (g *GoogleHandler) HandleGoogleOauthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	if code == "" {
		log.Print("code is empty")
		http.Error(w, "code is empty", http.StatusBadRequest)
		return
	}

	tok, err := g.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "token exchange failed", http.StatusBadRequest)
		return
	}

	log.Printf("code=%s", code)
	log.Printf("request=%#v", r)
	if token_id := tok.Extra("token_id"); token_id != nil {
		token := tok.AccessToken

		db := NewDB("orgo", "orgo.db")
		defer db.Close()
		if err := db.Put([]byte(token_id.(string)), []byte(token)); err != nil {
			log.Print(err.Error())
			http.Error(w, "save token", http.StatusBadRequest)
			return
		}
		log.Printf("google=%#v", tok)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

// HandleVerifyIdentity checks token_id against TokenInfo service to validate
// expiration and signature if the token is available in the session
func (g *GoogleHandler) HandleVerifyToken(w http.ResponseWriter, r *http.Request) {
	tokenID := r.FormValue("token_id")

	// Write the session id and redirect to /
	client := g.oauthConfig.Client(context.Background(), &oauth2.Token{})
	service, err := oauth2api.New(client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tokenCall := service.Tokeninfo()
	tokenCall.IdToken(tokenID)
	tokenInfo, err := tokenCall.Do()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("%#v", tokenInfo)
	http.Error(w, "", http.StatusOK)
}
