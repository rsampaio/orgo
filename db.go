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

func (d *DB) SaveToken(provider, account, code, token string) {

}

func (d *DB) GetToken(provider string) string {
	return ""
}

func (d *DB) Close() error {
	return d.handle.Close()
}
