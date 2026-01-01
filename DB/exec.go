package DB

import (
	"database/sql"
)

func (db *ConnDB) Exec(sqlStr string, arg ...any) (r sql.Result, err error) {
	if db == nil {
		return nil, ErrNoConn
	}
	args := handleNull(arg...)
	query := Replace(sqlStr, "?", db.Sign)
	db.log("Exec", query, args...).logDEBUG()
	if r, err = db.DBFunc.Conn.Exec(query, args...); err == nil {
		return
	}
	db.log("exec error", query, args...).logERROR(err)
	role := db.Conf.Role
	if opErr := db.checkOpError(err); opErr {
		if db = GetNewPool(role); db == nil {
			return nil, ErrNoConn
		}
	}
	if r, err = db.DBFunc.Conn.Exec(query, args...); err == nil {
		return
	}
	db.log("exec error", query, args...).logERROR(err)
	return
}

func (db *ConnDB) Affected(sqlStr string, arg ...any) (Affected int64, err error) {
	if db == nil {
		return 0, ErrNoConn
	}
	args := handleNull(arg)
	query := Replace(sqlStr, "?", db.Sign)
	db.log("Affected", query, args...).logDEBUG()
	var result sql.Result
	if result, err = db.DBFunc.Conn.Exec(query, args...); err == nil {
		return result.RowsAffected()
	}
	db.log("Affected error", query, args...).logERROR(err)
	role := db.Conf.Role
	if opErr := db.checkOpError(err); opErr {
		if db = GetNewPool(role); db == nil {
			return 0, ErrNoConn
		}
	}
	if result, err = db.DBFunc.Conn.Exec(query, args...); err == nil {
		return result.RowsAffected()
	}
	db.log("Affected error", query, args...).logERROR(err)
	return
}

func MysqlAddReturnId(mapper *Mapper) (insertId int64, err error) {
	db := mapper.Write()
	if db == nil {
		return 0, ErrNoConn
	}
	args := handleNull(mapper.Complete.Args...)
	var result sql.Result
	sqlStr := mapper.Complete.Sql + " RETURNING id"
	query := Replace(sqlStr, "?", db.Sign)
	db.log("MysqlAddReturnId", query, args...).logDEBUG()
	if result, err = db.DBFunc.Conn.Exec(query, args...); err == nil {
		return result.LastInsertId()
	}
	db.log("MysqlAddReturnId error", query, args...).logERROR(err)
	role := db.Conf.Role
	if opErr := db.checkOpError(err); opErr {
		if db = GetNewPool(role); db == nil {
			return 0, ErrNoConn
		}
	}
	if result, err = db.DBFunc.Conn.Exec(query, args...); err == nil {
		return result.LastInsertId()
	}
	db.log("MysqlAddReturnId error", query, args...).logERROR(err)
	return
}

func PgsqlAddReturnId(mapper *Mapper) (insertId int64, err error) {
	db := mapper.Write()
	if db == nil {
		return 0, ErrNoConn
	}
	args := handleNull(mapper.Complete.Args...)
	sqlStr := mapper.Complete.Sql + " RETURNING id"
	query := Replace(sqlStr, "?", db.Sign)
	db.log("PgsqlAddReturnId", query, args...).logDEBUG()
	if err = db.DBFunc.Conn.QueryRow(query, args...).Scan(&insertId); err == nil {
		return
	}
	db.log("PgsqlAddReturnId", query, args...).logERROR(err)
	role := db.Conf.Role
	if opErr := db.checkOpError(err); opErr {
		if db = GetNewPool(role); db == nil {
			return 0, ErrNoConn
		}
	}
	if err = db.DBFunc.Conn.QueryRow(query, args...).Scan(&insertId); err == nil {
		return
	}
	db.log("PgsqlAddReturnId", query, args...).logERROR(err)
	return
}

func MssqlAddReturnId(mapper *Mapper) (insertId int64, err error) {
	db := mapper.Write()
	if db == nil {
		return 0, ErrNoConn
	}
	args := handleNull(mapper.Complete.Args...)
	sqlStr := mapper.Complete.Sql + " ;SELECT SCOPE_IDENTITY();"
	query := Replace(sqlStr, "?", db.Sign)
	db.log("MssqlAddReturnId", query, args...).logDEBUG()
	if err = db.DBFunc.Conn.QueryRow(query, args...).Scan(&insertId); err == nil {
		return
	}
	db.log("MssqlAddReturnId error", query, args...).logERROR(err)
	role := db.Conf.Role
	if opErr := db.checkOpError(err); opErr {
		if db = GetNewPool(role); db == nil {
			return 0, ErrNoConn
		}
	}
	if err = db.DBFunc.Conn.QueryRow(query, args...).Scan(&insertId); err == nil {
		return
	}
	db.log("MssqlAddReturnId error", query, args...).logERROR(err)
	return
}

func (mapper *Mapper) ExecSql(sql string, args ...any) (r sql.Result, err error) {
	mapper.Complete.Sql = sql
	mapper.Complete.Args = args
	mapper.debug("ExecSql")
	if r, err = mapper.Write().Exec(mapper.Complete.Sql, mapper.Complete.Args...); err != nil {
		return
	}
	return
}

func (mapper *Mapper) Exec() (r sql.Result, err error) {
	if mapper.Complete.Sql, err = mapper.getSql(); err != nil {
		mapper.log("exec error").logERROR(err)
		return
	}
	mapper.debug("Exec")
	if r, err = mapper.Write().Exec(mapper.Complete.Sql, mapper.Complete.Args...); err != nil {
		return
	}
	return
}
