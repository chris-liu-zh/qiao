package cache

import (
	"database/sql"
	"fmt"
	"time"
)

const maxBatchSize = 1000 // 最大批量大小

type Store interface {
	createTable() error
	load() (map[string]Item, error)
	insert(key string, value []byte, expire int64) error
	delete(key string) error
	flush() error
	deleteExpire() error
	batchSet(items map[string]Item) error
}

type KVStore struct {
	db *sql.DB
}

func NewKVStore(dataSourceName string) (Store, error) {
	newKv := &KVStore{}
	var err error
	if newKv.db, err = sql.Open("sqlite3", dataSourceName); err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	if err = newKv.createTable(); err != nil {
		return nil, err
	}
	return newKv, nil
}

func (s *KVStore) createTable() error {
	// 如果不存在创建表
	if _, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS kv (
		  "key" TEXT NOT NULL,
		  "value" blob,
		  "expire" integer NOT NULL DEFAULT 0,
		  PRIMARY KEY ("key")
		)
	`); err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}
	return s.deleteExpire()
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
	if _, err := s.db.Exec(`INSERT OR REPLACE INTO kv (key, value, expire) VALUES (?, ?, ?)`, key, value, expire); err != nil {
		return fmt.Errorf("failed to insert: %v", err)
	}
	return nil
}

func (s *KVStore) delete(key string) error {
	if _, err := s.db.Exec(`DELETE FROM kv WHERE key = ?`, key); err != nil {
		return fmt.Errorf("failed to delete: %v", err)
	}
	return nil
}

func (s *KVStore) deleteExpire() error {
	if _, err := s.db.Exec(`DELETE FROM kv WHERE expire > 0 AND expire < ?`, time.Now().UnixNano()); err != nil {
		return fmt.Errorf("failed to delete expire: %v", err)
	}
	if _, err := s.db.Exec("VACUUM"); err != nil {
		return fmt.Errorf("failed to vacuum: %v", err)
	}
	return nil
}

func (s *KVStore) flush() error {
	if _, err := s.db.Exec(`DELETE FROM kv`); err != nil {
		return fmt.Errorf("failed to flush: %v", err)
	}
	if _, err := s.db.Exec("VACUUM"); err != nil {
		return fmt.Errorf("failed to vacuum: %v", err)
	}
	return nil
}

// batchSet 批量设置缓存项
func (s *KVStore) batchSet(items map[string]Item) error {
	// 超过最大批量大小，分块处理
	if len(items) > maxBatchSize {
		return s.batchSetInChunks(items, maxBatchSize)
	}
	// 开始事务
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	stmt, err := tx.Prepare(`INSERT OR REPLACE INTO kv (key, value, expire) VALUES (?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	// 执行批量插入
	for key, value := range items {
		if _, err := stmt.Exec(key, value.Object, value.Expiration); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute statement for key %s: %v", key, err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (s *KVStore) batchSetInChunks(items map[string]Item, chunkSize int) error {
	chunk := make(map[string]Item, chunkSize)
	i := 0
	for k, v := range items {
		chunk[k] = v
		i++
		if i%chunkSize == 0 {
			if err := s.batchSet(chunk); err != nil {
				return err
			}
			chunk = make(map[string]Item, chunkSize)
		}
	}
	// 处理剩余部分
	if len(chunk) > 0 {
		return s.batchSet(chunk)
	}
	return nil
}
