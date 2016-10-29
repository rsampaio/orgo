package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/sessions"
	"github.com/uber-go/zap"
)

type WebHandler struct {
	ctx    context.Context
	logger zap.Logger
	store  *sessions.CookieStore
}

func NewWebHandler(ctx context.Context, store *sessions.CookieStore) *WebHandler {
	return &WebHandler{logger: zap.New(zap.NewTextEncoder()), ctx: ctx, store: store}
}

// HandleIndex wrap requests to protected resources
func (h *WebHandler) IndexMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := h.store.Get(r, "orgo-session")
		if ok := session.Values["session_id"]; ok != nil {
			h.logger.Info("session id", zap.String("session_id", session.Values["session_id"].(string)))
			r.URL.Path = "/logged.html"
		} else {
			r.URL.Path = "/"
		}

		if r.URL.Path == "/" {
			r.URL.Path = "/index.html"
		}

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

	tmpl, err := template.ParseFiles(layout, bodyTmpl)
	if err != nil {
		h.logger.Error(err.Error())
		http.Error(w, "template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", templateOptions); err != nil {
		h.logger.Error(err.Error())
		http.Error(w, "template", http.StatusInternalServerError)
	}
}
