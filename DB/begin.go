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
	db     *ConnDB
	stmt   *sql.Stmt
	Err    error
	Mapper *Mapper
}

// Begin 开始事务
func (mapper *Mapper) Begin() *Begin {
	begin := &Begin{
		Mapper: mapper,
	}
	db := mapper.Write()
	if db == nil {
		begin.Err = ErrNoConn
		return begin
	}
	begin.db = db
	if begin.Tx, begin.Err = db.DBFunc.Conn.Begin(); begin.Err == nil {
		return begin
	}
	return begin
}

func (begin *Begin) Prepare(sqlStr string) *Begin {
	if begin.Err != nil {
		return begin
	}
	query := Replace(sqlStr, "?", begin.db.Sign)
	begin.db.log("Prepare exec", query).logDEBUG()
	begin.stmt, begin.Err = begin.Tx.Prepare(query)
	return begin
}

func (begin *Begin) StmtExec(args ...any) *Begin {
	if begin.Err != nil {
		return begin
	}
	txArgs := handleNull(args...)
	if _, begin.Err = begin.stmt.Exec(txArgs...); begin.Err != nil {
		return begin
	}
	return begin
}

func (begin *Begin) Exec(sqlStr string, args ...any) *Begin {
	if begin.Err != nil {
		return begin
	}
	txArgs := handleNull(args...)
	query := Replace(sqlStr, "?", begin.db.Sign)
	begin.db.log("Begin exec", query, args...).logDEBUG()
	if _, begin.Err = begin.Tx.Exec(query, txArgs...); begin.Err != nil {
		return begin
	}
	return begin
}

func (begin *Begin) Rollback() (err error) {
	if err = begin.Tx.Rollback(); err != nil {
		return err
	}
	return
}

func (begin *Begin) Commit() (err error) {
	if err = begin.Err; err != nil {
		return err
	}
	if err = begin.Tx.Commit(); err != nil {
		return err
	}
	return
}
