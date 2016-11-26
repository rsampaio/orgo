package main

import (
	"database/sql"
	"io/ioutil"

	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
	"github.com/uber-go/zap"
)

type DB struct {
	handle *sql.DB
	logger zap.Logger
}

func NewDB(file string) *DB {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil
	}
	return &DB{handle: db, logger: zap.New(zap.NewTextEncoder())}
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

func (d *DB) SaveSession(userID string) (string, error) {
	sessionID := uuid.NewV4().String()
	tx, err := d.handle.Begin()
	if err != nil {
		d.logger.Error(err.Error())
		return "", err
	}

	stmt, err := tx.Prepare("insert into sessions(sid, account) values(?, ?)")
	defer stmt.Close()
	if err != nil {
		d.logger.Error(err.Error())
		return "", err
	}

	if _, err := stmt.Exec(sessionID, userID); err != nil {
		tx.Rollback()
		d.logger.Error("save session", zap.String("error", err.Error()))
		return "", err
	}
	tx.Commit()
	return sessionID, nil
}

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

func (d *DB) GetDropboxId(googleID string) (string, error) {
	var dropboxID string

	err := d.handle.QueryRow("select dropbox_id from map_google_dropbox where google_id=?", googleID).Scan(&dropboxID)
	if err != nil {
		d.logger.Error("get dropbox id", zap.String("error", err.Error()))
		return "", err
	}

	return dropboxID, nil
}

func (d *DB) GetSession(sessionID string) (string, error) {
	var userID string

	err := d.handle.QueryRow("select account from sessions where sid=?", sessionID).Scan(&userID)
	if err != nil {
		d.logger.Error(err.Error())
		return "", err
	}

	return userID, nil
}

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

func (d *DB) GetToken(provider, account string) (string, error) {
	var token string
	err := d.handle.QueryRow("select token from tokens where provider=? and account=?", provider, account).Scan(&token)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (d *DB) Close() error {
	return d.handle.Close()
}
