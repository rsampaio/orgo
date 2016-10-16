package main

import (
	"context"
	"net/http"

	"github.com/joeshaw/envdecode"
	"github.com/uber-go/zap"
)

var logger zap.Logger

func main() {
	logger = zap.New(zap.NewTextEncoder())

	var cfg Config
	var urls map[string]string
	err := envdecode.Decode(&cfg)
	if err != nil {
		logger.Fatal(err.Error())
	}

	ctx := context.Background()

	// Buffered work chan for async producers
	workChan := make(chan string, 100)
	go WaitWork(workChan)

	dropboxHandler := NewDropboxHandler(cfg.Dropbox.ApiKey, cfg.Dropbox.ApiSecret, cfg.Dropbox.RedirectURL, workChan)
	googleHandler := NewGoogleHandler(cfg.Google.ApiKey, cfg.Google.ApiSecret, cfg.Google.RedirectURL)

	urls = map[string]string{
		"Dropbox": dropboxHandler.AuthCodeURL(),
		"Google":  googleHandler.AuthCodeURL(),
	}
	ctx = context.WithValue(ctx, "Urls", urls)
	webHandler := NewWebHandler(ctx, cfg.HttpCookieSecret)

	// Default handler
	http.HandleFunc("/dropbox/webhook", dropboxHandler.HandleWebhook)
	http.HandleFunc("/dropbox/oauth", dropboxHandler.HandleOauthCallback)
	http.HandleFunc("/google/oauth", googleHandler.HandleGoogleOauthCallback)
	http.HandleFunc("/google/verify_token", googleHandler.HandleVerifyToken)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", webHandler.HandleTemplates)

	logger.Fatal(http.ListenAndServe(":8080", nil).Error())
}
