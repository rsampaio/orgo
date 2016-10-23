package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"

	"github.com/uber-go/zap"
)

type WebHandler struct {
	ctx    context.Context
	logger zap.Logger
}

func NewWebHandler(ctx context.Context) *WebHandler {
	return &WebHandler{logger: zap.New(zap.NewTextEncoder()), ctx: ctx}
}

// HandleIndex requests to protected resources
func (h *WebHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		r.URL.Path = "/index.html"
	}
	h.HandleTemplates(w, r)
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
