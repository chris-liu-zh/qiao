/*
 * @Author: Chris
 * @Date: 2023-06-08 10:04:50
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-05 18:51:57
 * @Description: 请填写简介
 */
package DB

import (
	"database/sql"
)

var Update = "Update ${table} set ${set} ${where} ${order} ${group}"

// 更新数据并返回应向行数
func (mapper *Mapper) UpdateAffected(set any, args ...any) (affected int64, err error) {
	mapper = mapper.Set(set, args...)
	mapper.SqlTpl = Update
	if mapper.Complete.Sql, err = mapper.getSql(); err != nil {
		mapper.log("get sql error").logERROR(err)
		return
	}
	mapper.debug("UpdateAffected")
	if affected, err = mapper.Write().Affected(mapper.Complete.Sql, mapper.Complete.Args...); err != nil {
		return
	}
	return
}

// 更新数据并返回sql.Result
func (mapper *Mapper) Update(data, params any, args ...any) (r sql.Result, err error) {
	mapper = mapper.Set(data)
	mapper = mapper.Find(params, args...)
	mapper.SqlTpl = Update
	if mapper.Complete.Sql, err = mapper.getSql(); err != nil {
		mapper.log("get sql error").logERROR(err)
		return
	}
	mapper.debug("Update")
	if r, err = mapper.Write().Exec(mapper.Complete.Sql, mapper.Complete.Args...); err != nil {
		return
	}
	return
}
