package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

type Event struct {
	ListFolder struct {
		Accounts []string `json:"accounts"`
	} `json:"list_folder"`

	Delta struct {
		Users []int `json:"users"`
	} `json:"delta"`
}

type SignatureMismatchErr struct{}

func (s *SignatureMismatchErr) Error() string {
	return "SignatureMismatchError"
}

type DropboxHandler struct {
	OauthConfig *oauth2.Config
	workChan    chan string
}

func NewDropboxHandler(apiKey string, apiSecret string, redirectURL string, workChan chan string) *DropboxHandler {
	oauthConfig := &oauth2.Config{
		ClientID:     apiKey,
		ClientSecret: apiSecret,
		RedirectURL:  redirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.dropbox.com/1/oauth2/authorize",
			TokenURL: "https://api.dropbox.com/1/oauth2/token",
		},
	}
	return &DropboxHandler{OauthConfig: oauthConfig, workChan: workChan}
}

func (h *DropboxHandler) verifyRequest(r *http.Request) (bytes.Buffer, error) {
	var buf bytes.Buffer
	signature := r.Header.Get("X-Dropbox-Signature")
	mac := hmac.New(sha256.New, []byte(h.OauthConfig.ClientSecret))
	io.Copy(mac, io.TeeReader(r.Body, &buf))
	expectedMac := mac.Sum(nil)
	actualMac, err := hex.DecodeString(signature)
	if err != nil {
		log.Printf("decode signature error: %s", err.Error())
		return bytes.Buffer{}, &SignatureMismatchErr{}
	}

	if !hmac.Equal(actualMac, expectedMac) {
		return bytes.Buffer{}, &SignatureMismatchErr{}
	}
	return buf, nil
}

func (h *DropboxHandler) HandleOauthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	tok, err := h.OauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "token exchange failed", http.StatusBadRequest)
	}

	uid := tok.Extra("account_id").(string)
	token := tok.AccessToken
	db := NewDB("orgo", "orgo.db")
	if err := db.Put([]byte(uid), []byte(token)); err != nil {
		log.Print(err.Error())
		http.Error(w, "save token", http.StatusBadRequest)
	}

	db.Close()
	log.Printf("uid=%s account_id=%s\n", tok.Extra("uid"), tok.Extra("account_id"))
}

func (h *DropboxHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		log.Println("challenge request")
		fmt.Fprint(w, r.FormValue("challenge"))
	case "POST":
		if body, err := h.verifyRequest(r); err == nil {
			var event Event
			err := json.NewDecoder(&body).Decode(&event)
			if err != nil {
				log.Printf("decoder error: %s", err.Error())
			}

			for _, account := range event.ListFolder.Accounts {
				//go Process(account)
				h.workChan <- account
			}
		}

		http.Error(w, "", http.StatusOK)
	}
}
