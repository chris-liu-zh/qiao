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
	"time"
)

// 检测数据库错误是否为网络错误,并使用重连机制
func (db *ConnDB) check(err error) bool {
	if _, ok := err.(*net.OpError); ok {
		//网络错误，断开连接
		db.IsClose = true
		go db.reconnect() //异步重连
		return true
	}
	return false
}

func (db *ConnDB) reconnect() (err error) {
	//TODO 重连机制
	for range Pool.ReconnectNum {
		if db.DBFunc.Conn, err = db.connect(); err != nil {
			time.Sleep(Pool.ReconnectInterval)
			db.log("reconnect error", db.Conf.Dsn).logERROR(err)
			continue
		}
		db.IsClose = false
		db.log("reconnect success", db.Conf.Dsn).logINFO()
		return
	}
	return
}

func (db *ConnDB) checkOnline() bool {
	if db == nil {
		return false
	}
	if err := db.DBFunc.Conn.Ping(); err != nil {
		db.IsClose = true
		db.log(err.Error(), "").logWARNING()
		return false
	}
	return true
}
