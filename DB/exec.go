package DB

import (
	"database/sql"
)

func (db *ConnDB) Exec(sqlStr string, arg ...any) (r sql.Result, err error) {
	if db == nil {
		return nil, ErrNoConn()
	}
	args := handleNull(arg...)
	query := Replace(sqlStr, "?", db.Sign)
	db.log("Exec", query, args...).logDEBUG()
	if r, err = db.DBFunc.Conn.Exec(query, args...); err == nil {
		return
	}
	db.log(err.Error(), query, args...).logERROR()
	var ok bool
	if db, ok = online(db); ok {
		return
	}

	if db = GetNewPool(db.Part); db == nil {
		return nil, ErrNoConn()
	}
	if r, err = db.DBFunc.Conn.Exec(query, args...); err == nil {
		return
	}
	db.log(err.Error(), query, args...).logERROR()
	return
}

func (db *ConnDB) Affected(sqlStr string, arg ...any) (Affected int64, err error) {
	if db == nil {
		return 0, ErrNoConn()
	}
	args := handleNull(arg)
	query := Replace(sqlStr, "?", db.Sign)
	db.log("Affected", query, args...).logDEBUG()
	var result sql.Result
	if result, err = db.DBFunc.Conn.Exec(query, args...); err == nil {
		return result.RowsAffected()
	}
	db.log(err.Error(), query, args...).logERROR()
	var ok bool
	if db, ok = online(db); ok {
		return
	}

	if db = GetNewPool(db.Part); db == nil {
		return 0, ErrNoConn()
	}
	if result, err = db.DBFunc.Conn.Exec(query, args...); err == nil {
		return result.RowsAffected()
	}
	db.log(err.Error(), query, args...).logERROR()
	return
}

func MysqlAddReturnId(db *ConnDB, sqlStr string, arg ...any) (insertId int64, err error) {
	if db == nil {
		return 0, ErrNoConn()
	}
	args := handleNull(arg...)
	var result sql.Result
	query := Replace(sqlStr, "?", db.Sign)
	db.log("MysqlAddReturnId", query, args...).logDEBUG()
	if result, err = db.DBFunc.Conn.Exec(query, args...); err == nil {
		return result.LastInsertId()
	}
	db.log(err.Error(), query, args...).logERROR()
	var ok bool
	if db, ok = online(db); ok {
		return
	}

	if db = GetNewPool(db.Part); db == nil {
		return 0, ErrNoConn()
	}
	if result, err = db.DBFunc.Conn.Exec(query, args...); err == nil {
		return result.LastInsertId()
	}
	db.log(err.Error(), query, args...).logERROR()
	return
}

func PgsqlAddReturnId(db *ConnDB, sqlStr string, arg ...any) (insertId int64, err error) {
	if db == nil {
		return 0, ErrNoConn()
	}
	args := handleNull(arg...)
	sqlStr = sqlStr + " RETURNING id"
	query := Replace(sqlStr, "?", db.Sign)
	db.log("PgsqlAddReturnId", query, args...).logDEBUG()
	if err = db.DBFunc.Conn.QueryRow(query, args...).Scan(&insertId); err == nil {
		return
	}
	db.log(err.Error(), query, args...).logERROR()
	var ok bool
	if db, ok = online(db); ok {
		return
	}

	if db = GetNewPool(db.Part); db == nil {
		return 0, ErrNoConn()
	}
	if err = db.DBFunc.Conn.QueryRow(query, args...).Scan(&insertId); err == nil {
		return
	}
	db.log(err.Error(), query, args...).logERROR()
	return
}

func MssqlAddReturnId(db *ConnDB, sqlStr string, arg ...any) (insertId int64, err error) {
	if db == nil {
		return 0, ErrNoConn()
	}
	args := handleNull(arg...)
	sqlStr = sqlStr + " ;SELECT SCOPE_IDENTITY();"
	query := Replace(sqlStr, "?", db.Sign)
	db.log("MssqlAddReturnId", query, args...).logDEBUG()
	if err = db.DBFunc.Conn.QueryRow(query, args...).Scan(&insertId); err == nil {
		return
	}
	db.log(err.Error(), query, args...).logERROR()
	var ok bool
	if db, ok = online(db); ok {
		return
	}

	if db = GetNewPool(db.Part); db == nil {
		return 0, ErrNoConn()
	}
	if err = db.DBFunc.Conn.QueryRow(query, args...).Scan(&insertId); err == nil {
		return
	}
	db.log(err.Error(), query, args...).logERROR()
	return
}

func (mapper *Mapper) ExecSql(sql string) (r sql.Result, err error) {
	mapper.Complete.Sql = sql
	mapper.debug("ExecSql")
	if r, err = Write().Exec(mapper.Complete.Sql, mapper.Complete.Args...); err != nil {
		return
	}
	return
}

func (mapper *Mapper) Exec() (r sql.Result, err error) {
	if mapper.Complete.Sql, err = mapper.getSql(); err != nil {
		mapper.log(err.Error()).logERROR()
		return
	}
	mapper.debug("Exec")
	if r, err = Write().Exec(mapper.Complete.Sql, mapper.Complete.Args...); err != nil {
		return
	}
	return
}
