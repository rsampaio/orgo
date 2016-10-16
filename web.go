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
	store  *sessions.CookieStore
	logger zap.Logger
}

func NewWebHandler(ctx context.Context, cookieSecret string) *WebHandler {
	store := sessions.NewCookieStore([]byte(cookieSecret))
	return &WebHandler{logger: zap.New(zap.NewTextEncoder()), ctx: ctx, store: store}
}

func (g *WebHandler) HandleTemplates(w http.ResponseWriter, r *http.Request) {
	var templateOptions interface{}
	layout := path.Join("tmpl", "layout.html")

	// if not authenticated load index
	if r.URL.Path == "/" {
		r.URL.Path = "/index.html"
	}

	if r.URL.Path == "/index.html" {
		templateOptions = g.ctx.Value("Urls")
	}

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
		g.logger.Error(err.Error())
		http.Error(w, "template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", templateOptions); err != nil {
		g.logger.Error(err.Error())
		http.Error(w, "template", http.StatusInternalServerError)
	}
}
