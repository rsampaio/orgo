package main

import (
	"context"
	"net/http"

	log "github.com/Sirupsen/logrus"
	oauth2google "golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
	oauth2api "google.golang.org/api/oauth2/v2"

	"github.com/gorilla/sessions"
	"github.com/joeshaw/envdecode"
	"github.com/rsampaio/orgo/conf"
	"github.com/rsampaio/orgo/dropbox"
	"github.com/rsampaio/orgo/google"
	"github.com/rsampaio/orgo/web"
	"github.com/rsampaio/orgo/work"
	"golang.org/x/oauth2"
)

func main() {
	var (
		cfg  conf.Config
		urls map[string]string
	)

	err := envdecode.Decode(&cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	ctx := context.Background()
	store := sessions.NewCookieStore([]byte(cfg.HTTPCookieSecret))

	googleOauth := &oauth2.Config{
		ClientID:     cfg.Google.APIKey,
		ClientSecret: cfg.Google.APISecret,
		RedirectURL:  cfg.Google.RedirectURL,
		Endpoint:     oauth2google.Endpoint,
		Scopes:       []string{calendar.CalendarScope, oauth2api.UserinfoEmailScope, oauth2api.UserinfoProfileScope},
	}

	dropboxOauth := &oauth2.Config{
		ClientID:     cfg.Dropbox.APIKey,
		ClientSecret: cfg.Dropbox.APISecret,
		RedirectURL:  cfg.Dropbox.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.dropbox.com/1/oauth2/authorize",
			TokenURL: "https://api.dropbox.com/1/oauth2/token",
		},
	}

	// Buffered work chan for async producers
	worker := work.NewWorker(googleOauth, dropboxOauth)

	go worker.WaitWork()

	dropboxHandler := dropbox.NewDropboxHandler(dropboxOauth, worker.WorkChan, store)
	googleHandler := google.NewGoogleHandler(googleOauth, store)

	urls = map[string]string{
		"Dropbox": dropboxHandler.AuthCodeURL(),
		"Google":  googleHandler.AuthCodeURL(),
	}

	handler := web.NewHandler(ctx, store, urls)

	// Default handler
	http.HandleFunc("/dropbox/webhook", dropboxHandler.WebhookHandler)
	http.HandleFunc("/dropbox/oauth", dropboxHandler.OauthHandler)
	http.HandleFunc("/google/oauth", googleHandler.OauthHandler)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	templateHandler := http.HandlerFunc(handler.TemplateHandler)
	http.Handle("/", handler.IndexMiddleware(templateHandler))

	log.Fatal(http.ListenAndServe(":8080", nil).Error())
}
