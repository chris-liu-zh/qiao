package tools

import (
	"time"
)

func NowDateTime() string {
	return time.Now().Format(time.DateTime)
}

// TimeSecond 时间戳（秒）
func TimeSecond() int64 {
	return time.Now().Unix()
}

// TimeStamp 时间戳（毫秒）
func TimeStamp() int64 {
	return time.Now().UnixNano() / 1e6
}

// TimeNano 时间戳（纳秒）
func TimeNano() int64 {
	return time.Now().UnixNano()
}

// DateTimeStamp 日期时间转时间截
func DateTimeStamp(data string) (timestamp int64, err error) {
	return DateToStamp(time.DateTime, data)
}

// DateToStamp 时间转时间截
func DateToStamp(datetpl, data string) (timestamp int64, err error) {
	times, err := time.Parse(datetpl, data)
	if err != nil {
		return
	}
	timestamp = times.Unix()
	return
}

// StampToDateStamp 时间截转整（天）
func StampToDateStamp(timestamp time.Time) time.Time {
	hms := int64(timestamp.Second()) + int64((60 * timestamp.Minute())) + int64(timestamp.Hour()*3600)
	birthDay := timestamp.Unix() - hms
	return time.Unix(birthDay, 0)
}

// StampToDate 时间截转日期
func StampToDate(datetpl string, timestamp int64) string {
	timeobj := time.Unix(timestamp, 0)
	return timeobj.Format(datetpl)
}

// BeforeTimestamp 获取以前时间截
func BeforeTimestamp(y, m, d int) time.Time {
	return time.Now().AddDate(y, m, d)
}

// BeforeDate 获取以前日期
func BeforeDate(y, m, d int) string {
	return BeforeTimestamp(y, m, d).Format(time.DateTime)
}
