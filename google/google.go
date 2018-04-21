package google

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	orgodb "github.com/rsampaio/orgo/db"
	uuid "github.com/satori/go.uuid"
	oauth2api "google.golang.org/api/oauth2/v2"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

// GoogleHandler struct with unexported fields
type GoogleHandler struct {
	oauthConfig *oauth2.Config
	store       *sessions.CookieStore
	db          *orgodb.DB
}

// NewGoogleHandler creates an instance of GoogleHandler
func NewGoogleHandler(oauth *oauth2.Config, store *sessions.CookieStore) *GoogleHandler {
	return &GoogleHandler{
		store:       store,
		oauthConfig: oauth,
		db:          orgodb.NewDB("orgo.db"),
	}
}

// AuthCodeURL returns the URL to get the authentication code
func (g *GoogleHandler) AuthCodeURL() string {
	return g.oauthConfig.AuthCodeURL(uuid.NewV4().String(), oauth2.AccessTypeOffline)
}

// ServiceGetter get a google service configured
func (g *GoogleHandler) ServiceGetter(userID string) {
	t, err := g.db.GetToken("google", userID)
	log.Info(t.AccessToken, t.Code, err)
}

// OauthHandler will receive the AuthCode from when the user authorize the app
// with the code we should store the token_id and the code for future use and validation
func (g *GoogleHandler) OauthHandler(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	if code == "" {
		log.Info("code is empty")
		http.Error(w, "code is empty", http.StatusBadRequest)
		return
	}

	tok, err := g.oauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "token exchange failed", http.StatusBadRequest)
		return
	}

	client := g.oauthConfig.Client(oauth2.NoContext, &oauth2.Token{AccessToken: tok.AccessToken})
	service, _ := oauth2api.New(client)

	tokenCall := service.Tokeninfo()
	tokenCall.AccessToken(tok.AccessToken)
	tokenInfo, _ := tokenCall.Do()

	g.db.SaveToken("google", tokenInfo.UserId, code, tok)

	// Session
	session, err := g.store.Get(r, "orgo-session")
	if err != nil {
		log.Error(err.Error())
	}

	sessionID, _ := g.db.SaveSession(tokenInfo.UserId)
	session.Values["session_id"] = sessionID
	session.Save(r, w)
	http.Error(w, "ok", http.StatusOK)
}
