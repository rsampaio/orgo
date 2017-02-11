package dropbox

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

func TestDropbox(t *testing.T) {
	var (
		workChan = make(chan string, 10)
		store    = sessions.NewCookieStore([]byte("test-secret"))

		dropboxOauth = &oauth2.Config{
			ClientID:     "client-id",
			ClientSecret: "apiSecret123",
			RedirectURL:  "http://localhost",
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://www.dropbox.com/1/oauth2/authorize",
				TokenURL: "https://api.dropbox.com/1/oauth2/token",
			},
		}
		testHandler = NewDropboxHandler(dropboxOauth, workChan, store)
	)

	t.Run("dropbox_webhook_handler", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(testHandler.WebhookHandler))
		defer ts.Close()

		res, err := http.Get(fmt.Sprintf("%s?challenge=challenge-value", ts.URL))
		if err != nil {
			t.Errorf("shit went wrong %s", err.Error())
		}

		if res.Request.URL.Query().Get("challenge") != "challenge-value" {
			t.Errorf("challenge invalid %#q", res.Request.URL.Query())
		}

		body := `{"list_folder": {"accounts": ["account1"]}, "delta":{"users": [123]}}`

		client := http.Client{}
		req, err := http.NewRequest("POST", ts.URL, strings.NewReader(body))
		if err != nil {
			t.Errorf("client error %q", err.Error())
		}

		// Create hmac using secret
		mac := hmac.New(sha256.New, []byte("apiSecret123"))
		// Write data
		io.Copy(mac, strings.NewReader(body))
		// Checksum
		actualMac := hex.EncodeToString(mac.Sum(nil))
		fmt.Printf("test mac: %s\n", actualMac)

		req.Header.Add("X-Dropbox-Signature", actualMac)

		res, err = client.Do(req)
		if err != nil {
			t.Errorf("post error %q", err.Error())
		}

		if res.StatusCode != http.StatusOK {
			t.Errorf("webhook post status: %q", res.Status)
		}
	})

	t.Run("dropbox_oauth_handler", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(testHandler.OauthHandler))
		defer ts.Close()
	})
}
