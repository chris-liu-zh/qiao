/*
 * @Author: Chris
 * @Date: 2023-06-08 09:34:08
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-05 18:50:46
 * @Description: 请填写简介
 */
package DB

import (
	"database/sql"
)

func (db *ConnDB) QueryRow(sqlStr string, args ...any) (row *sql.Row) {
	if db == nil {
		return nil
	}
	query := Replace(sqlStr, "?", db.Sign)
	db.log("QueryRow", query, args).logDEBUG()
	if row = db.DBFunc.Conn.QueryRow(query, args...); row.Err() == nil {
		return
	}
	db.log(row.Err().Error(), query, args).logERROR()
	var ok bool
	if db, ok = online(db); ok {
		return
	}

	if db = GetNewPool(db.Part); db == nil {
		return nil
	}

	if row = db.DBFunc.Conn.QueryRow(query, args...); row.Err() == nil {
		return
	}
	db.log(row.Err().Error(), query, args).logERROR()
	return
}

func (db *ConnDB) Query(sqlStr string, args ...any) (rows *sql.Rows, err error) {
	if db == nil {
		return nil, ErrNoConn()
	}
	query := Replace(sqlStr, "?", db.Sign)
	db.log("Query", query, args).logDEBUG()
	if rows, err = db.DBFunc.Conn.Query(query, args...); err == nil {
		return
	}
	db.log(err.Error(), query, args).logERROR()
	var ok bool
	if db, ok = online(db); ok {
		return
	}
	if db = GetNewPool(db.Part); db == nil {
		return nil, ErrNoConn()
	}
	if rows, err = db.DBFunc.Conn.Query(query, args...); err == nil {
		return
	}
	db.log(err.Error(), query, args).logERROR()
	return
}

func (db *ConnDB) Count(sqlStr string, args ...any) (RowsCount int, err error) {
	if db == nil {
		return 0, ErrNoConn()
	}
	RowsCount = 0
	query := Replace(sqlStr, "?", db.Sign)
	db.log("Count", query, args).logDEBUG()
	if err = db.DBFunc.Conn.QueryRow(query, args...).Scan(&RowsCount); err == nil {
		return
	}
	db.log(err.Error(), query, args).logERROR()
	var ok bool
	if db, ok = online(db); ok {
		return
	}
	if db = GetNewPool(db.Part); db == nil {
		return 0, ErrNoConn()
	}
	if err = db.DBFunc.Conn.QueryRow(query, args...).Scan(&RowsCount); err == nil {
		return
	}
	db.log(err.Error(), query, args).logERROR()
	return
}

// Query 直接查询sql语句
func (mapper *Mapper) Query(sql string, args ...any) (rows *sql.Rows, err error) {
	mapper.Complete = SqlComplete{Sql: sql, Args: args}
	if rows, err = Read().Query(sql, args...); err != nil {
		mapper.log(err.Error()).logERROR()
		return
	}
	mapper.debug("Query")
	return
}

// QueryRow 直接查询sql语句
func (mapper *Mapper) QueryRow(sql string, args ...any) (row *sql.Row) {
	mapper.Complete = SqlComplete{Sql: sql, Args: args}
	if row = Read().QueryRow(sql, args...); row.Err() != nil {
		mapper.log(row.Err().Error()).logERROR()
		return
	}
	mapper.debug("QueryRow")
	return
}
