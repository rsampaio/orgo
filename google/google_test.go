package google

import (
	"testing"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

func TestGoogle(t *testing.T) {
	var (
		testHandler *GoogleHandler
		store       = sessions.NewCookieStore([]byte("test-secret"))
	)

	t.Run("new_google_handler", func(t *testing.T) {
		googleOauth := &oauth2.Config{
			ClientID:     "client-id",
			ClientSecret: "apiSecret123",
			RedirectURL:  "http://localhost",
		}
		testHandler = NewGoogleHandler(googleOauth, store)
	})

	t.Run("auth_code_url", func(t *testing.T) {
		if testHandler.AuthCodeURL() == "" {
			t.Error("auth_code url invalid")
		}
	})
}
