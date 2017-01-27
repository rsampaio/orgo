package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/sessions"
)

type WebHandler struct {
	ctx   context.Context
	store *sessions.CookieStore
	db    *DB
}

func NewWebHandler(ctx context.Context, store *sessions.CookieStore) *WebHandler {
	return &WebHandler{ctx: ctx, store: store, db: NewDB("orgo.db")}
}

// HandleIndex wrap requests to protected resources
func (h *WebHandler) IndexMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := h.store.Get(r, "orgo-session")
		if ok := session.Values["session_id"]; ok != nil {
			log.Info("session", session.Values["session_id"].(string))
			userID, err := h.db.GetSession(session.Values["session_id"].(string))
			if err != nil {
				r.URL.Path = "/error.html"
				goto reply
			}

			log.Info("session", userID)
			_, err = h.db.GetDropboxId(userID)
			if err != nil {
				r.URL.Path = "/dropbox.html"
				goto reply
			} else {
				r.URL.Path = "/logged.html"
				goto reply
			}
		} else {
			r.URL.Path = "/"
		}

		if r.URL.Path == "/" {
			r.URL.Path = "/index.html"
		}
	reply:
		next.ServeHTTP(w, r)
	})
}

func (h *WebHandler) HandleTemplates(w http.ResponseWriter, r *http.Request) {
	var templateOptions interface{}
	layout := path.Join("tmpl", "layout.html")
	bodyTmpl := path.Join("tmpl", r.URL.Path)

	info, err := os.Stat(bodyTmpl)

	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("not found: %s", r.URL.Path), http.StatusNotFound)
			return
		}
	}

	if info.IsDir() {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	templateOptions = h.ctx.Value("Urls").(map[string]string)

	tmpl, err := template.ParseFiles(layout, bodyTmpl)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", templateOptions); err != nil {
		log.Error(err.Error())
		http.Error(w, "template", http.StatusInternalServerError)
	}
}
