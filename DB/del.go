/*
 * @Author: Chris
 * @Date: 2023-06-08 10:04:34
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-05 18:48:05
 * @Description: 请填写简介
 */
package DB

import "database/sql"

var Del = "DELETE FROM ${table} ${where} ${order} ${group}"

// 删除数据
func (mapper *Mapper) Del() (r sql.Result, err error) {
	mapper.SqlTpl = Del
	if mapper.Complete.Sql, err = mapper.getSql(); err != nil {
		mapper.log(err.Error()).logERROR()
		return
	}
	mapper.debug("Del")
	if r, err = Write().Exec(mapper.Complete.Sql, mapper.Complete.Args...); err != nil {
		return
	}
	return
}

/*
删除数据并返回应向行数
*/
func (mapper *Mapper) DelAffected() (affected int64, err error) {
	mapper.SqlTpl = Del
	if mapper.Complete.Sql, err = mapper.getSql(); err != nil {
		mapper.log(err.Error()).logERROR()
		return
	}
	mapper.debug("DelAffected")
	if affected, err = Write().Affected(mapper.Complete.Sql, mapper.Complete.Args...); err != nil {
		return
	}
	return
}
