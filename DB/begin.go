/*
 * @Author: Chris
 * @Date: 2024-05-16 22:38:04
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-05 18:47:50
 * @Description: 请填写简介
 */
package DB

import (
	"database/sql"
)

type Begin struct {
	Tx     *sql.Tx
	Sign   string
	Title  string
	Part   string
	Err    error
	Mapper *Mapper
}

func (tx *Begin) Add(data any) *Begin {
	if tx.Err != nil {
		return tx
	}
	tx.Mapper.Debris = SqlDebris{}
	tx.Mapper.Complete = SqlComplete{Debug: tx.Mapper.Complete.Debug}
	tx.Mapper = tx.Mapper.getInsert(data)
	if tx = tx.Exec(); tx.Err != nil {
		return tx
	}
	return tx
}

func (db *ConnDB) Begin() (begin *Begin) {
	begin = &Begin{}
	if db == nil {
		begin.Err = ErrNoConn()
		return
	}

	begin.Sign = db.Sign
	begin.Title = db.Title
	begin.Part = db.Part
	if begin.Tx, begin.Err = db.DBFunc.Conn.Begin(); begin.Err == nil {
		return
	}
	var ok bool
	if db, ok = online(db); ok {
		return begin
	}

	if db = GetNewPool(db.Part); db == nil {
		begin.Err = ErrNoConn()
		return begin
	}
	if begin.Tx, begin.Err = db.DBFunc.Conn.Begin(); begin.Err == nil {
		return
	}
	return
}

func (tx *Begin) Exec() *Begin {
	if tx.Err != nil {
		return tx
	}
	if tx.Mapper.Complete.Sql, tx.Err = tx.Mapper.getSql(); tx.Err != nil {
		return tx
	}
	args := handleNull(tx.Mapper.Complete.Args...)
	query := Replace(tx.Mapper.Complete.Sql, "?", tx.Sign)
	if _, tx.Err = tx.Tx.Exec(query, args...); tx.Err != nil {
		if tx.Err = tx.Tx.Rollback(); tx.Err != nil {
			return tx
		}
		return tx
	}
	return tx
}

func (tx *Begin) Commit() error {
	tx.Mapper.debug("Commit")
	if tx.Err != nil {
		tx.log(tx.Err.Error()).logERROR()
		return tx.Err
	}
	return tx.Tx.Commit()
}
