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

	"github.com/chris-liu-zh/qiao/tools"
)

type Begin struct {
	Tx     *sql.Tx
	stmt   *sql.Stmt
	Err    error
	Title  string
	Mapper *Mapper
}

// Begin 开始事务
func (mapper *Mapper) Begin() *Begin {
	tx := &Begin{
		Mapper: mapper,
	}
	db := mapper.Write()
	if db == nil {
		tx.Err = ErrNoConn
		return tx
	}
	tx.Title = db.Conf.Title
	if tx.Tx, tx.Err = db.DBFunc.Conn.Begin(); tx.Err == nil {
		return tx
	}
	return tx
}

func (tx *Begin) Prepare(sqlStr string) *Begin {
	if tx.Err != nil {
		return tx
	}
	tx.stmt, tx.Err = tx.Tx.Prepare(sqlStr)
	return tx
}

func (tx *Begin) StmtExec(args ...any) *Begin {
	if tx.Err != nil {
		return tx
	}
	txArgs := handleNull(args...)
	if _, tx.Err = tx.stmt.Exec(txArgs...); tx.Err != nil {
		return tx
	}
	return tx
}

func (tx *Begin) Exec(args ...any) *Begin {
	if tx.Err != nil {
		return tx
	}
	if tx.Mapper.Complete.Sql, tx.Err = tx.Mapper.getSql(); tx.Err != nil {
		return tx
	}
	txArgs := handleNull(args...)
	query := Replace(tx.Mapper.Complete.Sql, "?", tx.Mapper.Debris.sign)
	if _, tx.Err = tx.Tx.Exec(query, txArgs...); tx.Err != nil {
		return tx
	}
	return tx
}

func (tx *Begin) Rollback() (err error) {
	defer tools.DeferErr(&err, tx.stmt.Close)
	if err = tx.Tx.Rollback(); err != nil {
		return err
	}
	return
}

func (tx *Begin) Commit() (err error) {
	defer tools.DeferErr(&err, tx.stmt.Close)
	if err = tx.Err; err != nil {
		return err
	}
	if err = tx.Tx.Commit(); err != nil {
		return err
	}
	return
}
