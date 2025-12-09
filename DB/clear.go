/*
 * @Author: Chris
 * @Date: 2023-06-07 17:34:08
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-05 18:12:18
 * @Description: 请填写简介
 */
package DB

import (
	"net"
	"slices"
	"time"
)

// 检测数据库错误是否为网络错误,并使用重连机制
func (db *ConnDB) check(err error) bool {
	if _, ok := err.(*net.OpError); ok {
		//网络错误，断开连接
		db.IsClose = true
		clear(db.Conf.Role) //清理连接池
		go db.reconnect()   //异步重连
		return true
	}
	return false
}

func (db *ConnDB) reconnect() {
	//TODO 重连机制
	for range Pool.ReconnectNum {
		time.Sleep(Pool.ReconnectInterval)
		if err := db.Conf.NewDB(); err != nil {
			db.log("reconnect error", db.Conf.Dsn).logERROR(err)
			continue
		}
		return
	}
}

func checkOnline(db *ConnDB) (*ConnDB, bool) {
	if err := db.DBFunc.Conn.Ping(); err != nil {
		db.IsClose = true
		db.log(err.Error(), "").logWARNING()
		go clear(db.Conf.Role)
		return nil, false
	}
	return db, true
}

func clear(role string) {
	switch role {
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
