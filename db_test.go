package main

import "testing"

func TestDB(t *testing.T) {
	db := NewDB("")
	if db == nil || db.handle == nil {
		t.Error("failed to open db")
	}

	t.Run("create_tables", func(t *testing.T) {
		err := db.createTables()
		if err != nil {
			t.Error(err.Error())
		}
	})

	t.Run("session_save_get", func(t *testing.T) {
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

	t.Run("google_dropbox_map", func(t *testing.T) {
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

	t.Run("token", func(t *testing.T) {
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

	db.Close()
}
