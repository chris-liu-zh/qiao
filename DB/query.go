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

func (db *ConnDB) Query(sqlStr string, args ...any) (rows *sql.Rows, err error) {
	if db == nil {
		return nil, ErrNoConn
	}
	query := Replace(sqlStr, "?", db.Sign)
	db.log("Query", query, args).logDEBUG()
	if rows, err = db.DBFunc.Conn.Query(query, args...); err == nil {
		return
	}
	db.log("Query error", query, args).logERROR(err)
	role := db.Conf.Role
	if opErr := db.checkOpError(err); opErr {
		if db = GetNewPool(role); db == nil {
			return nil, ErrNoConn
		}
	}
	if rows, err = db.DBFunc.Conn.Query(query, args...); err == nil {
		return
	}
	db.log("Query error", query, args).logERROR(err)
	return
}

func (db *ConnDB) Count(sqlStr string, args ...any) (RowsCount int, err error) {
	if db == nil {
		return 0, ErrNoConn
	}
	RowsCount = 0
	query := Replace(sqlStr, "?", db.Sign)
	db.log("Count", query, args).logDEBUG()
	if err = db.DBFunc.Conn.QueryRow(query, args...).Scan(&RowsCount); err == nil {
		return
	}
	db.log("Count error", query, args).logERROR(err)
	role := db.Conf.Role
	if opErr := db.checkOpError(err); opErr {
		if db = GetNewPool(role); db == nil {
			return 0, ErrNoConn
		}
	}
	if err = db.DBFunc.Conn.QueryRow(query, args...).Scan(&RowsCount); err == nil {
		return
	}
	db.log("Count error", query, args).logERROR(err)
	return
}

// Query 直接查询sql语句
func (mapper *Mapper) Query(sql string, args ...any) (*Mapper, error) {
	mapper.Complete = SqlComplete{Sql: sql, Args: args}
	var err error
	if mapper.sqlRows, err = mapper.Read().Query(sql, args...); err != nil {
		mapper.log("Query error").logERROR(err)
		return nil, err
	}
	mapper.debug("Query")
	return mapper, nil
}

// QueryRow 直接查询sql语句
func (mapper *Mapper) QueryRow(sql string, args ...any) (*Mapper, error) {
	var err error
	mapper.Complete = SqlComplete{Sql: sql, Args: args}
	if mapper.sqlRows, err = mapper.Read().Query(sql, args...); err != nil {
		mapper.log("Query error").logERROR(err)
		return nil, err
	}
	mapper.debug("QueryRow")
	return mapper, nil
}
