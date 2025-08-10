package qiao

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"
)

func UUIDV7() string {
	var uuid [16]byte
	// 1. 获取当前 Unix 时间戳（毫秒）
	now := time.Now()
	unixMilli := now.UnixMilli()
	// 2. 写入前 48 位（时间戳）
	binary.BigEndian.PutUint64(uuid[0:8], uint64(unixMilli)<<16) // 高 48 位
	// 3. 写入版本号 `7`（第 6-7 字节的第 4-7 位）
	uuid[6] = 0x70 | (uuid[6] & 0x0F) // 0111 0000
	// 4. 写入变体 `10`（第 8 字节的高 2 位）
	uuid[8] = 0x80 | (uuid[8] & 0x3F) // 1000 0000
	// 5. 填充剩余 62 位随机数
	_, err := rand.Read(uuid[8:16])
	if err != nil {
		return ""
	}

	// 6. 格式化为标准 UUID 字符串
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(uuid[0:4]),
		binary.BigEndian.Uint16(uuid[4:6]),
		binary.BigEndian.Uint16(uuid[6:8]),
		binary.BigEndian.Uint16(uuid[8:10]),
		uuid[10:16],
	)
}
