package db

import (
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	db "upper.io/db.v3"
)

type Token struct {
	TokenType    string    `db:"token_type"`
	AccessToken  string    `db:"token"`
	RefreshToken string    `db:"token_refresh"`
	Expiry       time.Time `db:"expiry"`
	Code         string    `db:"code"`
	Provider     string    `db:"provider"`
	Account      string    `db:"account"`
}

type MapGoogleDropbox struct {
	GoogleID  string `db:"google_id"`
	DropboxID string `db:"dropbox_id"`
}

// SaveGoogleDropbox creates the map between google and dropbox accounts.
func (d *DB) SaveGoogleDropbox(googleID, dropboxID string) error {
	col := d.sess.Collection("map_google_dropbox")
	_, err := col.Insert(&MapGoogleDropbox{GoogleID: googleID, DropboxID: dropboxID})
	return err
}

// GetDropboxID retrieves the dropbox id for a given google id.
func (d *DB) GetDropboxID(googleID string) (string, error) {
	var result MapGoogleDropbox
	if err := d.sess.Collection("map_google_dropbox").Find(db.Cond{"google_id": googleID}).One(&result); err != nil {
		return "", err
	}
	return result.DropboxID, nil
}

func (d *DB) GetGoogleID(dropboxID string) (string, error) {
	var result MapGoogleDropbox
	err := d.sess.Collection("map_google_dropbox").Find(db.Cond{"dropbox_id": dropboxID}).One(&result)
	return result.GoogleID, err
}

// SaveToken saves the code from an OAUTH nepotiation with a provider for a specific account.
func (d *DB) SaveToken(provider, account, code string, token *oauth2.Token) error {
	col := d.sess.Collection("tokens")
	_, err := col.Insert(&Token{
		Account:      account,
		Provider:     provider,
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		Code:         code,
	})
	return err
}

// GetToken retrieves the token for a provider and account.
func (d *DB) GetToken(provider, account string) (Token, error) {
	var result Token
	err := d.sess.Collection("tokens").Find(db.Cond{"provider": provider}, db.Cond{"account": account}).One(&result)
	return result, errors.Wrap(err, "get token")
}
