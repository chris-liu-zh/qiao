package ignore

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
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

func (c *cache) writeDataRecord(key string, value []byte, epx time.Time) (int64, error) {
	entry := &walEntry{
		Op:        opInsert,
		Timestamp: epx.UnixNano(),
		Key:       []byte(key),
		Value:     value,
		KeySize:   uint32(len(key)),
		ValueSize: uint32(len(value)),
	}
	// 获取当前文件偏移量
	offset, err := c.dataFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	// 写入记录头
	header := make([]byte, 16)
	binary.BigEndian.PutUint64(header[0:8], uint64(entry.Timestamp))
	binary.BigEndian.PutUint32(header[8:12], entry.KeySize)
	binary.BigEndian.PutUint32(header[12:16], entry.ValueSize)
	if _, err := c.dataFile.Write(header); err != nil {
		return 0, err
	}

	// 写入key
	if _, err := c.dataFile.Write(entry.Key); err != nil {
		return 0, err
	}

	// 写入value
	if _, err := c.dataFile.Write(entry.Value); err != nil {
		return 0, err
	}

	return offset, nil
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

// Flush 清除缓存中的所有项目
func (c *cache) Flush() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.datas = map[string]Item{}
	if c.dataFile != nil {
		if err := os.Truncate(c.dataFile.Name(), 0); err != nil {
			return fmt.Errorf("error clearing cache file: %v", err)
		}
	}
	return nil
}
