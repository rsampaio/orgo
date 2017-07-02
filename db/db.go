package db

import (
	"io/ioutil"
	"time"

	upper "upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/sqlite"
)

// DB struct with unexported fields.
type DB struct {
	sess sqlbuilder.Database
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
	UserID    string    `db:"user_id"`
	Title     string    `db:"title"`
	Tag       string    `db:"tag"`
	Priority  string    `db:"priority"`
	Body      string    `db:"body"`
	Date      time.Time `db:"created_at"`
	Scheduled time.Time `db:"scheduled"`
	Closed    time.Time `db:"closed"`
}

// NewDB creates a new DB instance.
func NewDB(file string) *DB {
	db, err := sqlite.Open(&sqlite.ConnectionURL{Database: file})
	if err != nil {
		return nil
	}
	return &DB{sess: db}
}

// GetEntry retrieves an OrgEntry from the database by title and email.
func (d *DB) GetEntry(title, userID string) (*OrgEntry, error) {
	var entry OrgEntry
	err := d.sess.Collection("entries").Find(upper.Cond{"title": title}, upper.Cond{"user_id": userID}).One(&entry)
	return &entry, err
}

// SaveEntry saves an OrgEntry to the database.
func (d *DB) SaveEntry(entry *OrgEntry) error {
	col := d.sess.Collection("entries")
	_, err := col.Insert(entry)
	return err
}

func (d *DB) SaveOrUpdate(entry *OrgEntry) error {
	_, err := d.GetEntry(entry.Title, entry.UserID)
	if err != nil {
		err := d.SaveEntry(entry)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close closes database sessr
func (d *DB) Close() error {
	return d.sess.Close()
}

func (d *DB) createTables() error {
	sql, err := ioutil.ReadFile("sql/create_tables.sql")
	if err != nil {
		return err
	}

	_, err = d.sess.Exec(string(sql))
	if err != nil {
		return err
	}

	return nil
}
