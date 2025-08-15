package cache

import (
	"time"
)

type walEntry struct {
	Op        byte   // 操作类型
	Timestamp int64  // 时间戳
	KeySize   uint32 // 键的大小
	ValueSize uint32 // 值的大小
	Key       []byte // 键
	Value     []byte // 值
	Checksum  uint32 // 校验和
}

// 定期保存缓存
func (c *cache) startSaving() {
	if c.saveInterval >= 0 {
		ticker := time.NewTicker(c.saveInterval)
		go func() {
			for range ticker.C {
				if c.writeTotal > c.writeInterval {
					//if err := c.SaveFile(c.cachePath); err != nil {
					//	log.Printf("Error saving cache to file: %v\n", err)
					//	continue
					//}
					c.writeTotal = 0
				}
			}
		}()
	}
}
