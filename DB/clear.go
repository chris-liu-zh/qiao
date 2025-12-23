/*
 * @Author: Chris
 * @Date: 2023-06-07 17:34:08
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-05 18:12:18
 * @Description: 请填写简介
 */
package DB

import (
	"context"
	"errors"
	"net"
	"time"
)

// 检测数据库错误是否为网络错误,并使用重连机制
func (db *ConnDB) checkOpError(err error) bool {
	var opError *net.OpError
	if errors.As(err, &opError) {
		//网络错误，断开连接
		db.IsClose = true
		go db.reconnect() //异步重连
		return true
	}
	return false
}

func (db *ConnDB) reconnect() {
	if db == nil {
		return
	}
	//TODO 重连机制
	for range Pool.ReconnectNum {
		if db.IsClose {
			return
		}
		if ok := db.checkOnline(); ok {
			db.log("reconnect success", db.Conf.Dsn).logINFO()
			return
		}
		time.Sleep(Pool.ReconnectInterval)
	}
}

func (db *ConnDB) checkOnline() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.DBFunc.Conn.PingContext(ctx); err != nil {
		db.IsClose = true
		db.log("ping error", "").logERROR(err)
		return false
	}
	db.IsClose = false
	return true
}
