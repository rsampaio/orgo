package db

import (
	db "upper.io/db.v3"

	"github.com/google/uuid"
)

type Session struct {
	ID      string `db:"sid"`
	Account string `db:"account"`
}

// SaveSession creates an uuid ID for the given userID as a session.
func (d *DB) SaveSession(userID string) (string, error) {
	sessionID := uuid.New().String()
	sessionCollection := d.sess.Collection("sessions")
	if _, err := sessionCollection.Insert(&Session{ID: sessionID, Account: userID}); err != nil {
		return "", err
	}
	return sessionID, nil
}

// GetSession retrieves the userID for the sessionID provided.
func (d *DB) GetSession(sessionID string) (string, error) {
	var sessions []Session
	err := d.sess.Collection("sessions").Find(db.Cond{"sid": sessionID}).All(&sessions)
	if err != nil {
		return "", err
	}

	return sessions[0].Account, nil
}
