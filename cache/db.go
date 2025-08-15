package cache

import (
	"database/sql"
	"fmt"
	"time"
)

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

	return &KVStore{
		db: db,
	}, nil
}

func (s *KVStore) Close() {
	s.db.Close()
}
func (s *KVStore) Query() (row map[string]any, err error) {
	rows, err := s.db.Query(`SELECT key,value FROM kv`)
	if err != nil {
		return
	}
	defer rows.Close()
	if columns, err = rows.Columns(); err != nil {
		return
	}
	length := len(columns)
	pointer = make([]any, length)
	for i := range length {
		var val any
		pointer[i] = &val
	}
	for Rows.Next() {
		row := make(map[string]any)
		if err = Rows.Scan(pointer...); err == nil {
			for i := range length {
				row[columns[i]] = *pointer[i].(*any)
			}
			list = append(list, row)
		}
	}
	return
}

func (s *KVStore) Insert(key string, value any, expire time.Time) error {
	_, err := s.db.Exec(`INSERT INTO kv (key, value, expire) VALUES (?, ?, ?)`, key, value, expire.UnixNano())
	if err != nil {
		return fmt.Errorf("failed to insert: %v", err)
	}
	return nil
}
