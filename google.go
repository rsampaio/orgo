package main

import (
	"net/http"

	"github.com/gorilla/sessions"
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
	store       *sessions.CookieStore
	db          *DB
}

func NewGoogleHandler(apiKey, apiSecret, redirectURL string, store *sessions.CookieStore) *GoogleHandler {
	return &GoogleHandler{
		store:  store,
		logger: zap.New(zap.NewTextEncoder()),
		oauthConfig: &oauth2.Config{
			ClientID:     apiKey,
			ClientSecret: apiSecret,
			RedirectURL:  redirectURL,
			Endpoint:     google.Endpoint,
			Scopes:       []string{calendar.CalendarScope, oauth2api.UserinfoEmailScope, oauth2api.UserinfoProfileScope},
		},
		db: NewDB("orgo.db"),
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

	// TODO: save google token
	g.db.SaveToken("google", tokenInfo.UserId, code, tok.AccessToken)

	// Session
	session, err := g.store.Get(r, "orgo-session")
	if err != nil {
		g.logger.Error(err.Error())
	}

	sessionID, _ := g.db.SaveSession(tokenInfo.UserId)
	session.Values["session_id"] = sessionID
	session.Save(r, w)
	http.Error(w, "ok", http.StatusOK)
}
