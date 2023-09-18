package kv

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	bolt "go.etcd.io/bbolt"
)

// db is a wrapper around bolt.DB that keeps track of the number of references
// to the database and closes the database when the last reference is closed.
type db struct {
	handle   *bolt.DB
	opened   atomic.Bool
	refCount atomic.Int64
	lock     sync.Mutex
}

// newDB returns a new db instance.
func newDB() *db {
	return &db{
		handle:   new(bolt.DB),
		opened:   atomic.Bool{},
		refCount: atomic.Int64{},
		lock:     sync.Mutex{},
	}
}

// open opens the database if it is not already open.
//
// It is safe to call this method multiple times.
// The database will only be opened once.
func (db *db) open() error {
	if db.opened.Load() {
		db.refCount.Add(1)
		return nil
	}

	db.lock.Lock()
	defer db.lock.Unlock()

	if db.opened.Load() {
		return nil
	}

	handler, err := bolt.Open(DefaultKvPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	handler.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(DefaultKvBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	db.handle = handler
	db.opened.Store(true)
	db.refCount.Add(1)

	return nil
}

// close closes the database if there are no more references to it.
func (db *db) close() error {
	if db.refCount.Add(-1) == 0 {
		if err := db.handle.Close(); err != nil {
			return err
		}

		db.opened.Store(false)
	}

	return nil
}
