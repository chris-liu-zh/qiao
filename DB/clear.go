/*
 * @Author: Chris
 * @Date: 2023-06-07 17:34:08
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-05 18:12:18
 * @Description: 请填写简介
 */
package DB

import "slices"

func online(db *ConnDB) (*ConnDB, bool) {
	if err := db.DBFunc.Conn.Ping(); err != nil {
		db.IsClose = true
		db.log(err.Error(), "").logWARNING()
		go clear(db.Part)
		return nil, false
	}
	return db, true
}

func clear(part string) {
	switch part {
	case "master":
		poolList := Pool.Master.DBConn
		poolList = clearDB(poolList)
		Pool.Master.PoolNum = len(poolList)
	case "slave":
		poolList := Pool.Slave.DBConn
		poolList = clearDB(poolList)
		Pool.Slave.PoolNum = len(poolList)
	case "alone":
		poolList := Pool.Alone.DBConn
		poolList = clearDB(poolList)
		Pool.Alone.PoolNum = len(poolList)
	}
	Pool.PoolCount = Pool.Master.PoolNum + Pool.Slave.PoolNum + Pool.Alone.PoolNum
}

func clearDB(poolList []ConnDB) []ConnDB {
	for i := range poolList {
		if poolList[i].IsClose {
			if err := poolList[i].DBFunc.Conn.Close(); err != nil {
				poolList[i].log(err.Error(), "").logINFO()
			}
			poolList = slices.Delete(poolList, i, i+1)
			i--
		}
	}
	return poolList
}
