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
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/uber-go/zap"
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
	oauthConfig *oauth2.Config
	workChan    chan string
	logger      zap.Logger
	db          *DB
	store       *sessions.CookieStore
}

func NewDropboxHandler(apiKey string, apiSecret string, redirectURL string, workChan chan string, store *sessions.CookieStore) *DropboxHandler {
	oauthConfig := &oauth2.Config{
		ClientID:     apiKey,
		ClientSecret: apiSecret,
		RedirectURL:  redirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.dropbox.com/1/oauth2/authorize",
			TokenURL: "https://api.dropbox.com/1/oauth2/token",
		},
	}

	return &DropboxHandler{logger: zap.New(zap.NewTextEncoder()), oauthConfig: oauthConfig, workChan: workChan, db: NewDB("orgo.db"), store: store}
}

func (h *DropboxHandler) AuthCodeURL() string {
	return h.oauthConfig.AuthCodeURL("state-token", []oauth2.AuthCodeOption{}...)
}

func (h *DropboxHandler) verifyRequest(r *http.Request) (bytes.Buffer, error) {
	var buf bytes.Buffer
	signature := r.Header.Get("X-Dropbox-Signature")
	mac := hmac.New(sha256.New, []byte(h.oauthConfig.ClientSecret))
	io.Copy(mac, io.TeeReader(r.Body, &buf))
	expectedMac := mac.Sum(nil)
	actualMac, err := hex.DecodeString(signature)
	if err != nil {
		h.logger.Error("decode signature error", zap.String("msg", err.Error()))
		return bytes.Buffer{}, &SignatureMismatchErr{}
	}

	if !hmac.Equal(actualMac, expectedMac) {
		return bytes.Buffer{}, &SignatureMismatchErr{}
	}
	return buf, nil
}

func (h *DropboxHandler) HandleOauthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	tok, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "token exchange failed", http.StatusBadRequest)
	}

	uid := tok.Extra("account_id").(string)
	token := tok.AccessToken

	// save dropbox token
	h.db.SaveToken("dropbox", uid, code, token)
	session, err := h.store.Get(r, "orgo-session")
	if err != nil {
		h.logger.Error("get session", zap.String("error", err.Error()))
		return
	}

	sessionID := session.Values["session_id"].(string)
	userID, err := h.db.GetSession(sessionID)
	if err != nil {
		h.logger.Error(err.Error())
	}

	err = h.db.SaveGoogleDropbox(userID, tok.Extra("uid").(string))
	if err != nil {
		h.logger.Error("save google dropbox", zap.String("error", err.Error()))
		return
	}

	h.logger.Info("redirect",
		zap.String("uid", tok.Extra("uid").(string)),
		zap.String("session_id", sessionID),
		zap.String("google_id", userID),
		zap.String("account_id", tok.Extra("account_id").(string)))
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *DropboxHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.logger.Info("challenge request")
		fmt.Fprint(w, r.FormValue("challenge"))
	case "POST":
		if body, err := h.verifyRequest(r); err == nil {
			var event Event
			err := json.NewDecoder(&body).Decode(&event)
			if err != nil {
				h.logger.Info("decoder error", zap.String("msg", err.Error()))
			}

			for _, account := range event.ListFolder.Accounts {
				//go Process(account)
				h.workChan <- account
			}
		}

		http.Error(w, "", http.StatusOK)
	}
}
