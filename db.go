package main

import (
	"github.com/boltdb/bolt"
)

type DB struct {
	handle *bolt.DB
	bucket []byte
}

func NewDB(bucket string, file string) *DB {
	db, err := bolt.Open(file, 0666, nil)
	if err != nil {
		return nil
	}
	return &DB{handle: db, bucket: []byte(bucket)}
}

func (d *DB) Get(key []byte) ([]byte, error) {
	var result []byte
	err := d.handle.View(func(tx *bolt.Tx) error {
		result = tx.Bucket(d.bucket).Get(key)
		return nil
	})
	if err != nil {
		return []byte(""), err
	}

	return result, nil
}

func (d *DB) Put(key []byte, value []byte) error {
	return d.handle.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(d.bucket)
		if err != nil {
			return err
		}
		return b.Put(key, value)
	})
}

func (d *DB) Close() error {
	return d.handle.Close()
}
