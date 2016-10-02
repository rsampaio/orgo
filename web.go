package main

import "net/http"

type WebHandler struct {
}

func NewWebHandler() *WebHandler {
	return &WebHandler{}
}

func (g *WebHandler) HandleTemplates(w http.ResponseWriter, r *http.Request) {

}
