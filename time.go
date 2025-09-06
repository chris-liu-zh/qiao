/*
 * @Author: Chris
 * @Date: 2023-05-22 15:42:28
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-09 23:16:06
 * @Description: 请填写简介
 */
package qiao

import (
	"time"
)

// 当前时间的字符串，2006-01-02 15:04:05
const DateTime = "2006-01-02 15:04:05"
const DateDay = "2006-01-02"

func NowDateTime() string {
	return time.Now().Format(DateTime)
}

// 时间戳（秒）
func TimeSecond() int64 {
	return time.Now().Unix()
}

// 时间戳（毫秒）
func TimeStamp() int64 {
	return time.Now().UnixNano() / 1e6
}

// 时间戳（纳秒）
func TimeNano() int64 {
	return time.Now().UnixNano()
}

/**
 * @description:  日期时间转时间截
 * @param {*} datetpl
 * @param {string} data
 * @return {*}
 */
func DateTimeStamp(data string) (timestamp int64, err error) {
	times, err := time.Parse(DateTime, data)
	if err != nil {
		return
	}
	timestamp = times.Unix()
	return
}

/**
 * @description: 时间转时间截
 * @param {*} datetpl
 * @param {string} data
 * @return {*}
 */
func DateToStamp(datetpl, data string) (timestamp int64, err error) {
	times, err := time.Parse(datetpl, data)
	if err != nil {
		return
	}
	timestamp = times.Unix()
	return
}

/**
 * @description: 时间截转整（天）
 * @param {time.Time} timestamp
 * @return {*}
 */
func StampToDateStamp(timestamp time.Time) time.Time {
	hms := int64(timestamp.Second()) + int64((60 * timestamp.Minute())) + int64(timestamp.Hour()*3600)
	birthDay := timestamp.Unix() - hms
	return time.Unix(birthDay, 0)
}

/**
 * @description: 时间截转日期
 * @param {string} datetpl demo:"2006-01-02 15:04:05"
 * @param {int64} timestamp
 * @return {*}
 */
func StampToDate(datetpl string, timestamp int64) string {
	timeobj := time.Unix(timestamp, 0)
	return timeobj.Format(datetpl)
}

/**
 * @description:获取以前时间截
 * @param {*} y
 * @param {*} m
 * @param {int} d
 * @return {*}
 */
func BeforeTimestamp(y, m, d int) time.Time {
	return time.Now().AddDate(y, m, d)
}

/**
 * @description:获取以前日期
 * @param {*} y
 * @param {*} m
 * @param {int} d
 * @return {*}
 */
func BeforeDate(y, m, d int) string {
	return BeforeTimestamp(y, m, d).Format("2006-01-02 15:04:05")
}
