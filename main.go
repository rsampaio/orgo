package main

import (
	"log"
	"net/http"

	"github.com/joeshaw/envdecode"
)

func main() {
	var cfg Config
	err := envdecode.Decode(&cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Buffered work chan for async producers
	workChan := make(chan string, 100)
	go WaitWork(workChan)
	dropboxHandler := NewDropboxHandler(cfg.Dropbox.ApiKey, cfg.Dropbox.ApiSecret, cfg.Dropbox.RedirectURL, workChan)
	googleHandler := NewGoogleHandler()
	webHandler := NewWebHandler()

	// Default handler
	http.HandleFunc("/dropbox/webhook", dropboxHandler.HandleWebhook)
	http.HandleFunc("/dropbox/oauth", dropboxHandler.HandleOauthCallback)
	http.HandleFunc("/google/oauth", googleHandler.HandleGoogleOauthCallback)
	http.Handle("/static", http.FileServer(http.Dir("static")))
	http.Handle("/", webHandler.HandleTemplates)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
