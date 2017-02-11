package db

import (
	"database/sql"
	"io/ioutil"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
)

// DB struct with unexported fields.
type DB struct {
	handle *sql.DB
}

// OrgEntry struct defines an OrgMode entry in this format:
// ```
// ** [TAG [#PRIORITY]] Title
//    SCHEDULED:value CLOSED:value
//    body
//    body
//    [date]
// ```
type OrgEntry struct {
	UserID    string
	Title     string
	Tag       string
	Priority  string
	Body      []string
	Date      time.Time
	Scheduled time.Time
	Closed    time.Time
}

// NewDB creates a new DB instance.
func NewDB(file string) *DB {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil
	}
	return &DB{handle: db}
}

// SaveSession creates an uuid ID for the given userID as a session.
func (d *DB) SaveSession(userID string) (string, error) {
	sessionID := uuid.NewV4().String()
	tx, err := d.handle.Begin()
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	stmt, err := tx.Prepare("insert into sessions(sid, account) values(?, ?)")
	defer stmt.Close()
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	if _, err := stmt.Exec(sessionID, userID); err != nil {
		tx.Rollback()
		log.Error("save session", err.Error())
		return "", err
	}
	tx.Commit()
	return sessionID, nil
}

// GetSession retrieves the userID for the sessionID provided.
func (d *DB) GetSession(sessionID string) (string, error) {
	var userID string

	err := d.handle.QueryRow("select account from sessions where sid=?", sessionID).Scan(&userID)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	return userID, nil
}

// SaveGoogleDropbox creates the map between google and dropbox accounts.
func (d *DB) SaveGoogleDropbox(googleID, dropboxID string) error {
	tx, err := d.handle.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("insert into map_google_dropbox(google_id, dropbox_id) values(?, ?)")
	if err != nil {
		return err
	}

	if _, err = stmt.Exec(googleID, dropboxID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetDropboxID retrieves the dropbox id for a given google id.
func (d *DB) GetDropboxID(googleID string) (string, error) {
	var dropboxID string

	err := d.handle.QueryRow("select dropbox_id from map_google_dropbox where google_id=?", googleID).Scan(&dropboxID)
	if err != nil {
		log.Error("get dropbox id", err.Error())
		return "", err
	}

	return dropboxID, nil
}

func (d *DB) GetGoogleID(dropboxID string) (string, error) {
	var googleID string

	err := d.handle.QueryRow("select google_id from map_google_dropbox where dropbox_id=?", dropboxID).Scan(&googleID)
	if err != nil {
		log.Errorf("get google id: %v", err.Error())
		return "", err
	}

	return googleID, nil
}

// SaveToken saves the code from an OAUTH nepotiation with a provider for a specific account.
func (d *DB) SaveToken(provider, account, code, token string) error {
	tx, err := d.handle.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("insert into tokens(provider, account, code, token) values(?, ?, ?, ?)")
	defer stmt.Close()
	if err != nil {
		return err
	}

	if _, err := stmt.Exec(provider, account, code, token); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// GetToken retrieves the token for a provider and account.
func (d *DB) GetToken(provider, account string) (string, string, error) {
	var (
		token string
		code  string
	)
	err := d.handle.QueryRow("select token, code from tokens where provider=? and account=?", provider, account).Scan(&token, &code)
	if err != nil {
		return "", "", err
	}

	return token, code, nil
}

// GetEntry retrieves an OrgEntry from the database by title and email.
func (d *DB) GetEntry(title, userID string) (*OrgEntry, error) {
	var (
		entry OrgEntry
		body  string
	)

	err := d.handle.QueryRow(
		"select title, tag, priority, body, create_date, scheduled, closed from entries where title=? and userid=?",
		title,
		userID,
	).Scan(
		&entry.Title,
		&entry.Tag,
		&entry.Priority,
		&body,
		&entry.Date,
		&entry.Scheduled,
		&entry.Closed,
	)
	if err != nil {
		return nil, err
	}

	entry.Body = strings.Split(body, "\n")
	return &entry, nil
}

// SaveEntry saves an OrgEntry to the database.
func (d *DB) SaveEntry(entry *OrgEntry) error {
	tx, err := d.handle.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("insert into entries(userid, title, tag, priority, body, create_date, scheduled, closed) values(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}

	if _, err := stmt.Exec(
		entry.UserID,
		entry.Title,
		entry.Tag,
		entry.Priority,
		strings.Join(entry.Body, "\n"),
		entry.Date,
		entry.Scheduled,
		entry.Closed,
	); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// Close closes database handler
func (d *DB) Close() error {
	return d.handle.Close()
}

func (d *DB) createTables() error {
	sql, err := ioutil.ReadFile("sql/create_tables.sql")
	if err != nil {
		return err
	}

	_, err = d.handle.Exec(string(sql))
	if err != nil {
		return err
	}

	return nil
}
