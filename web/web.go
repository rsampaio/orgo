package web

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	orgodb "gitlab.com/rvaz/orgo/db"

	"github.com/gorilla/sessions"
)

// WebHandler struct with unexported fields.
type WebHandler struct {
	ctx   context.Context
	store *sessions.CookieStore
	urls  map[string]string
	db    *orgodb.DB
}

// NewWebHandler returns an instance of WebHandler.
func NewWebHandler(ctx context.Context, store *sessions.CookieStore, urls map[string]string) *WebHandler {
	return &WebHandler{
		ctx:   ctx,
		store: store,
		urls:  urls,
		db:    orgodb.NewDB("orgo.db"),
	}
}

// IndexMiddleware wrap requests to protected resources.
func (h *WebHandler) IndexMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := h.store.Get(r, "orgo-session")
		if ok := session.Values["session_id"]; ok != nil {
			log.Infof("session %s", session.Values["session_id"].(string))
			userID, err := h.db.GetSession(session.Values["session_id"].(string))
			if err != nil {
				r.URL.Path = "/error.html"
				goto reply
			}

			log.Infof("session %s", userID)
			_, err = h.db.GetDropboxID(userID)
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

// TemplateHandler render templates for specific urls.
func (h *WebHandler) TemplateHandler(w http.ResponseWriter, r *http.Request) {
	var (
		layout   = path.Join("tmpl", "layout.html")
		bodyTmpl = path.Join("tmpl", r.URL.Path)
	)

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

	tmpl, err := template.ParseFiles(layout, bodyTmpl)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", h.urls); err != nil {
		log.Error(err.Error())
		http.Error(w, "template", http.StatusInternalServerError)
	}
}
