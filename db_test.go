package main

import (
	"testing"
	"time"
)

func TestDB(t *testing.T) {
	db := NewDB("")
	defer db.Close()
	if db == nil || db.handle == nil {
		t.Error("failed to open db")
		return
	}

	err := db.createTables()
	if err != nil {
		t.Error(err.Error())
	}

	t.Run("SessionSaveGet", func(t *testing.T) {
		sessionID, err := db.SaveSession("user1")
		if err != nil {
			t.Error(err.Error())
		}

		u, err := db.GetSession(sessionID)
		if err != nil {
			t.Error(err.Error())
		}

		if u != "user1" {
			t.Error("session user invalid")
		}

	})

	t.Run("GoogleDropboxMap", func(t *testing.T) {
		err := db.SaveGoogleDropbox("google1", "dropbox1")
		if err != nil {
			t.Error(err.Error())
		}

		d, err := db.GetDropboxId("google1")
		if err != nil {
			t.Error(err.Error())
		}

		if d != "dropbox1" {
			t.Error("dropbox user map invalid")
		}
	})

	t.Run("Token", func(t *testing.T) {
		err := db.SaveToken("provider1", "account1", "abc123", "token123")
		if err != nil {
			t.Error(err.Error())
		}

		token, err := db.GetToken("provider1", "account1")
		if err != nil {
			t.Error(err.Error())
		}

		if token != "token123" {
			t.Error("token is invalid")
		}

	})

	t.Run("EntrySaveGet", func(t *testing.T) {
		var (
			ti    = time.Now()
			entry = &OrgEntry{
				Title:     "title",
				Tag:       "tag",
				Priority:  "prio",
				Body:      []string{"body"},
				Date:      ti,
				Scheduled: ti,
				Closed:    ti,
			}
		)

		if err := db.SaveEntry(entry, "test@email.com"); err != nil {
			t.Error(err.Error())
		}

		entry1, err := db.GetEntry("title", "test@email.com")
		if err != nil {
			t.Error(err.Error())
		}

		if entry1 == nil {
			t.Error("entry is nil")
		}
	})
}
