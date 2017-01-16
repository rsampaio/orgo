package main

import (
	"context"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/sessions"
	"github.com/joeshaw/envdecode"
)

func main() {
	var cfg Config
	var urls map[string]string
	err := envdecode.Decode(&cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	ctx := context.Background()
	store := sessions.NewCookieStore([]byte(cfg.HttpCookieSecret))

	// Buffered work chan for async producers
	workChan := make(chan string, 100)
	calendarChan := make(chan *OrgEntry)
	go WaitWork(workChan, calendarChan)

	dropboxHandler := NewDropboxHandler(cfg.Dropbox.ApiKey, cfg.Dropbox.ApiSecret, cfg.Dropbox.RedirectURL, workChan, store)
	googleHandler := NewGoogleHandler(cfg.Google.ApiKey, cfg.Google.ApiSecret, cfg.Google.RedirectURL, store)

	urls = map[string]string{
		"Dropbox": dropboxHandler.AuthCodeURL(),
		"Google":  googleHandler.AuthCodeURL(),
	}

	ctx = context.WithValue(ctx, "Urls", urls)
	webHandler := NewWebHandler(ctx, store)

	// Default handler
	http.HandleFunc("/dropbox/webhook", dropboxHandler.HandleWebhook)
	http.HandleFunc("/dropbox/oauth", dropboxHandler.HandleOauthCallback)
	http.HandleFunc("/google/oauth", googleHandler.HandleGoogleOauthCallback)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	templateHandler := http.HandlerFunc(webHandler.HandleTemplates)
	http.Handle("/", webHandler.IndexMiddleware(templateHandler))

	log.Fatal(http.ListenAndServe(":8080", nil).Error())
}
