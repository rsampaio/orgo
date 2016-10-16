package main

import (
	"context"
	"fmt"
	"net/http"

	uuid "github.com/satori/go.uuid"
	"github.com/uber-go/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
	oauth2api "google.golang.org/api/oauth2/v2"
)

type GoogleHandler struct {
	oauthConfig *oauth2.Config
	logger      zap.Logger
}

func NewGoogleHandler(apiKey, apiSecret, redirectURL string) *GoogleHandler {
	return &GoogleHandler{
		logger: zap.New(zap.NewTextEncoder()),
		oauthConfig: &oauth2.Config{
			ClientID:     apiKey,
			ClientSecret: apiSecret,
			RedirectURL:  redirectURL,
			Endpoint:     google.Endpoint,
			Scopes:       []string{calendar.CalendarScope, oauth2api.UserinfoEmailScope, oauth2api.UserinfoProfileScope},
		},
	}
}
func (g *GoogleHandler) AuthCodeURL() string {
	return g.oauthConfig.AuthCodeURL(uuid.NewV4().String(), oauth2.AccessTypeOffline)
}

// HandleGoogleOauthCallback will receive the AuthCode from when the user authorize the app
// with the code we should store the token_id and the code for future use and validation
func (g *GoogleHandler) HandleGoogleOauthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	if code == "" {
		g.logger.Info("code is empty")
		http.Error(w, "code is empty", http.StatusBadRequest)
		return
	}

	tok, err := g.oauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		g.logger.Error(err.Error())
		http.Error(w, "token exchange failed", http.StatusBadRequest)
		return
	}

	client := g.oauthConfig.Client(oauth2.NoContext, &oauth2.Token{AccessToken: tok.AccessToken})
	service, _ := oauth2api.New(client)
	tokenCall := service.Tokeninfo()
	tokenCall.AccessToken(tok.AccessToken)
	tokenInfo, _ := tokenCall.Do()

	g.logger.Info("google user id", zap.String("user_id", tokenInfo.UserId))

	db := NewDB("orgo", "orgo.db")
	defer db.Close()
	key := fmt.Sprintf("%s:access_token", tokenInfo.UserId)
	if err := db.Put([]byte(key), []byte(tok.AccessToken)); err != nil {
		g.logger.Error(err.Error())
		http.Error(w, "save token", http.StatusBadRequest)
		return
	}

	key = fmt.Sprintf("%s:code", tokenInfo.UserId)
	if err := db.Put([]byte(key), []byte(code)); err != nil {
		g.logger.Error(err.Error())
		http.Error(w, "save code", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// HandleVerifyIdentity checks token_id against TokenInfo service to validate
// expiration and signature if the token is available in the session
func (g *GoogleHandler) HandleVerifyToken(w http.ResponseWriter, r *http.Request) {
	tokenID := r.FormValue("access_token")

	// Write the session id and redirect to /
	client := g.oauthConfig.Client(context.Background(), &oauth2.Token{})
	service, err := oauth2api.New(client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tokenCall := service.Tokeninfo()
	tokenCall.AccessToken(tokenID)
	tokenInfo, err := tokenCall.Do()
	g.logger.Info(fmt.Sprintf("%#v", tokenInfo))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Error(w, "", http.StatusOK)
}
