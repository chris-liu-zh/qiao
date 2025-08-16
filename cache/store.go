package cache

import (
	"database/sql"
	"fmt"
	"time"
)

const maxBatchSize = 1000

type Store interface {
	load() (map[string]Item, error)
	insert(key string, value []byte, expire int64) error
	delete(key string) error
	flush() error
	deleteExpire() error
}

type KVStore struct {
	db *sql.DB
}

func NewKVStore(path string) (*KVStore, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// 如果不存在创建表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS kv (
		  "key" TEXT NOT NULL,
		  "value" blob,
		  "expire" integer NOT NULL DEFAULT 0,
		  PRIMARY KEY ("key")
		)
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	_, err = db.Exec(`DELETE FROM kv WHERE expire > 0 AND expire < ?`, time.Now().UnixNano())
	return &KVStore{
		db,
	}, err
}

func (s *KVStore) load() (items map[string]Item, err error) {
	rows, err := s.db.Query(`SELECT key,value,expire FROM kv where expire > ?`, time.Now().UnixNano())
	if err != nil {
		return
	}
	defer rows.Close()
	var key string
	var value []byte
	var expire int64
	items = make(map[string]Item)
	for rows.Next() {
		if err = rows.Scan(&key, &value, &expire); err == nil {
			items[key] = Item{
				Object:     value,
				Expiration: expire,
			}
		}
	}
	return
}

func (s *KVStore) insert(key string, value []byte, expire int64) error {
	_, err := s.db.Exec(`INSERT OR REPLACE INTO kv (key, value, expire) VALUES (?, ?, ?)`, key, value, expire)
	return err
}

func (s *KVStore) delete(key string) error {
	_, err := s.db.Exec(`DELETE FROM kv WHERE key = ?`, key)
	return err
}

func (s *KVStore) deleteExpire() error {
	_, err := s.db.Exec(`DELETE FROM kv WHERE expire > 0 AND expire < ?`, time.Now().UnixNano())
	return err
}

func (s *KVStore) flush() error {
	_, err := s.db.Exec(`DELETE FROM kv`)
	return err
}
