package cache

import (
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

const (
	maxBatchSize = 20000 // 最大批量大小
	delSql       = "DELETE FROM kv WHERE key IN (?)"
	putSql       = "INSERT OR REPLACE INTO kv (key,value,expire) VALUES (?,?,?)"
)

type Store interface {
	createTable() error
	load() (map[string]Item, error)
	flush() error
	deleteExpire() error
	put(key string, value []byte, expire int64) error
	delete(key string) error
	sync(cache *cache, sql string, opKeys []string) error
}

type KVStore struct {
	db  *sql.DB
	kvU sync.RWMutex
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
	s.kvU.Lock()
	defer s.kvU.Unlock()
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
		if err = rows.Scan(&key, &value, &expire); err != nil {
			slog.Error("failed to scan row", "err", err)
		}
		items[key] = Item{
			Object:     value,
			Expiration: expire,
		}
		continue
	}
	return
}

func (s *KVStore) put(key string, value []byte, expire int64) error {
	s.kvU.Lock()
	defer s.kvU.Unlock()
	_, err := s.db.Exec(`INSERT OR REPLACE INTO kv (key, value, expire) VALUES (?, ?, ?)`, key, value, expire)
	return err
}

func (s *KVStore) delete(key string) error {
	s.kvU.Lock()
	defer s.kvU.Unlock()
	_, err := s.db.Exec(`DELETE FROM kv WHERE key = ?`, key)
	return err
}

func (s *KVStore) deleteExpire() error {
	s.kvU.Lock()
	defer s.kvU.Unlock()
	if _, err := s.db.Exec(`DELETE FROM kv WHERE expire > 0 AND expire < ?`, time.Now().UnixNano()); err != nil {
		return fmt.Errorf("failed to delete expire: %v", err)
	}
	if _, err := s.db.Exec("VACUUM"); err != nil {
		return fmt.Errorf("failed to vacuum: %v", err)
	}
	return nil
}

func (s *KVStore) flush() error {
	s.kvU.Lock()
	defer s.kvU.Unlock()
	if _, err := s.db.Exec(`DELETE FROM kv`); err != nil {
		return fmt.Errorf("failed to flush: %v", err)
	}
	if _, err := s.db.Exec("VACUUM"); err != nil {
		return fmt.Errorf("failed to vacuum: %v", err)
	}
	return nil
}

func (s *KVStore) sync(c *cache, sql string, opKeys []string) (err error) {
	keyLen := len(opKeys)
	if keyLen == 0 {
		return
	}
	s.kvU.Lock()
	defer s.kvU.Unlock()
	// 超过最大批量大小，分块处理
	if keyLen > maxBatchSize {
		return s.batchSetInChunks(c, keyLen, sql, opKeys)
	}
	return s.batch(c, sql, opKeys)
}

func (s *KVStore) batch(c *cache, sql string, opKeys []string) (err error) {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
		if err = tx.Commit(); err != nil {
			slog.Error("failed to commit transaction:", "err", err)
		}
	}()
	stmt, err := tx.Prepare(sql)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	//执行批量删除
	if sql == delSql {
		for _, key := range opKeys {
			if _, err = stmt.Exec(key); err != nil {
				return fmt.Errorf("failed to execute statement: %v", err)
			}
		}
		return
	}
	// 执行批量插入
	if sql == putSql {
		for _, key := range opKeys {
			if _, err = stmt.Exec(key, c.items[key].Object, c.items[key].Expiration); err != nil {
				return fmt.Errorf("failed to execute statement: %v", err)
			}
		}
	}
	return nil
}

func (s *KVStore) batchSetInChunks(c *cache, length int, sql string, opKeys []string) error {
	// 超过最大批量大小，分块处理
	for i := 0; i < length; i += maxBatchSize {
		// 计算当前分组的结束索引
		end := min(i+maxBatchSize, length)
		// 执行批量操作
		if err := s.batch(c, sql, opKeys[i:end]); err != nil {
			return err
		}
	}
	return nil
}
