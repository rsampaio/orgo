package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	handle *sql.DB
}

func NewDB(file string) *DB {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil
	}
	return &DB{handle: db}
}

func (d *DB) createTables() error {
	_, err := d.handle.Exec(`
    create table user (
        email           text,
        token_id        integer
    );

    create table tokens (
        id       integer primary key autoincrement,
        provider text,
        account  text,
        code     text,
        token    text
    );

	create table events (
        email text,
        due   datetime,
        title text,
        state text,
        body  text);`)

	if err != nil {
		return err
	}

	return nil
}

func (d *DB) SaveToken(provider, account, code, token string) error {
	tx, err := d.handle.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("insert into tokens(provider, account, code, token) values(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(provider, account, code, token); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (d *DB) GetToken(provider string) (string, error) {
	var token string
	err := d.handle.QueryRow("select token from tokens where provider=?", provider).Scan(token)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (d *DB) Close() error {
	return d.handle.Close()
}
