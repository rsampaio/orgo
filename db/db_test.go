package db

import (
	"testing"
	"time"
)

func TestDB(t *testing.T) {
	d := NewDB("")
	defer d.Close()
	if d == nil || d.handle == nil {
		t.Fatal("failed to open db")
	}

	err := d.createTables()
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Run("SessionSaveGet", func(t *testing.T) {
		sessionID, err := d.SaveSession("user1")
		if err != nil {
			t.Fatal(err.Error())
		}

		u, err := d.GetSession(sessionID)
		if err != nil {
			t.Fatal(err.Error())
		}

		if u != "user1" {
			t.Fatal("session user invalid")
		}

	})

	t.Run("GoogleDropboxMap", func(t *testing.T) {
		err := d.SaveGoogleDropbox("google1", "dropbox1")
		if err != nil {
			t.Fatal(err.Error())
		}

		d, err := d.GetDropboxID("google1")
		if err != nil {
			t.Fatal(err.Error())
		}

		if d != "dropbox1" {
			t.Fatal("dropbox user map invalid")
		}
	})

	t.Run("Token", func(t *testing.T) {
		err := d.SaveToken("provider1", "account1", "abc123", "token123", "type", "refresh", time.Now())
		if err != nil {
			t.Fatal(err.Error())
		}

		t, err := d.GetToken("provider1", "account1")
		if err != nil {
			t.Fatal(err.Error())
		}

		if t.Token != "token123" {
			t.Fatal("token is invalid")
		}

		if t.Code != "abc123" {
			t.Fatalf("code is %v, want abc123", t.Code)
		}

	})

	t.Run("EntrySaveGet", func(t *testing.T) {
		var (
			ti    = time.Now()
			entry = &OrgEntry{
				UserID:    "test@email.com",
				Title:     "title",
				Tag:       "tag",
				Priority:  "prio",
				Body:      []string{"body"},
				Date:      ti,
				Scheduled: ti,
				Closed:    ti,
			}
		)

		err := d.SaveEntry(entry)
		if err != nil {
			t.Fatal(err.Error())
		}

		entry1, err := d.GetEntry("title", "test@email.com")
		if err != nil {
			t.Fatal(err.Error())
		}

		if entry1 == nil {
			t.Fatal("entry is nil")
		}
	})
}
