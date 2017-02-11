package dropbox

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	log "github.com/Sirupsen/logrus"
	orgodb "gitlab.com/rvaz/orgo/db"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

// Event struct encodes dropbox callback responses
type Event struct {
	ListFolder struct {
		Accounts []string `json:"accounts"`
	} `json:"list_folder"`

	Delta struct {
		Users []int `json:"users"`
	} `json:"delta"`
}

// ErrSignatureMismatch defines signature error
var ErrSignatureMismatch = errors.New("signature mismatch error")

// DropboxHandler struct with unexported fields
type DropboxHandler struct {
	oauthConfig *oauth2.Config
	workChan    chan string
	db          *orgodb.DB
	store       *sessions.CookieStore
}

// NewDropboxHandler returns a new DropboxHandler
func NewDropboxHandler(oauth *oauth2.Config, workChan chan string, store *sessions.CookieStore) *DropboxHandler {
	return &DropboxHandler{
		oauthConfig: oauth,
		workChan:    workChan,
		db:          orgodb.NewDB("orgo.db"),
		store:       store,
	}
}

// AuthCodeURL returns the url to redirect for an authorization code
func (h *DropboxHandler) AuthCodeURL() string {
	return h.oauthConfig.AuthCodeURL("state-token", []oauth2.AuthCodeOption{}...)
}

func (h *DropboxHandler) verifyRequest(r *http.Request) (bytes.Buffer, error) {
	var (
		buf       bytes.Buffer
		signature = r.Header.Get("X-Dropbox-Signature")
		mac       = hmac.New(sha256.New, []byte(h.oauthConfig.ClientSecret))
	)

	io.Copy(mac, io.TeeReader(r.Body, &buf))

	// Calculate expected mac after writing to mac
	expectedMac := mac.Sum(nil)

	actualMac, err := hex.DecodeString(signature)
	if err != nil {
		log.Errorf("decode signature error %s", err.Error())
		return bytes.Buffer{}, ErrSignatureMismatch
	}

	if !hmac.Equal(actualMac, expectedMac) {
		log.Errorf(
			"expected mac: %s got: %s",
			hex.EncodeToString(expectedMac),
			hex.EncodeToString(actualMac),
		)
		return bytes.Buffer{}, ErrSignatureMismatch
	}
	return buf, nil
}

// OauthHandler handles dropbox oauth calls
func (h *DropboxHandler) OauthHandler(w http.ResponseWriter, r *http.Request) {
	var (
		code = r.FormValue("code")
	)

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
		log.Errorf("get session %s", err.Error())
		return
	}

	sessionID := session.Values["session_id"].(string)
	userID, err := h.db.GetSession(sessionID)
	if err != nil {
		log.Error(err.Error())
	}

	err = h.db.SaveGoogleDropbox(userID, tok.Extra("account_id").(string))
	if err != nil {
		log.Errorf("save google dropbox %s", err.Error())
		return
	}

	log.Info("redirect ",
		tok.Extra("uid").(string),
		sessionID,
		userID,
		tok.Extra("account_id").(string))

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// WebhookHandler handles dropbok webhook events
func (h *DropboxHandler) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	var event Event

	switch r.Method {
	case "GET":
		log.Info("challenge request")
		fmt.Fprint(w, r.FormValue("challenge"))
	case "POST":
		body, err := h.verifyRequest(r)
		if err != nil {
			http.Error(w, "", http.StatusForbidden)
			return
		}

		err = json.NewDecoder(&body).Decode(&event)
		if err != nil {
			log.Infof("decoder error %s", err.Error())
		}

		for _, account := range event.ListFolder.Accounts {
			//go Process(account)
			h.workChan <- account
		}

		http.Error(w, "", http.StatusOK)
	}
}
