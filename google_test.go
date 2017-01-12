package main

import (
	"testing"

	"github.com/gorilla/sessions"
)

func TestGoogle(t *testing.T) {
	var (
		testHandler *GoogleHandler
		store       = sessions.NewCookieStore([]byte("test-secret"))
	)

	t.Run("new_google_handler", func(t *testing.T) {
		testHandler = NewGoogleHandler("api123", "apiSecret123", "http://localhost", store)
	})

	t.Run("auth_code_url", func(t *testing.T) {
		if testHandler.AuthCodeURL() == "" {
			t.Error("auth_code url invalid")
		}
	})
}
