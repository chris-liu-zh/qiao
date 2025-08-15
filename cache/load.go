package cache

import (
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
)

func (c *cache) Load(r io.Reader) error {
	dec := gob.NewDecoder(r)
	items := map[string]Item{}
	if err := dec.Decode(&items); err != nil && err != io.EOF {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("load cache file error: %s", err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range items {
		ov, found := c.items[k]
		if !found || ov.Expired() {
			c.items[k] = v
		}
	}
	return nil
}

func (c *cache) LoadFile() error {
	if _, err := c.dataFile.Seek(0, 0); err != nil {
		return err
	}

	var offset int64 = 0
	index := make(map[string]any)
	for {
		header := make([]byte, 16) // timestamp(8) + keySize(4) + valueSize(4)
		_, err := io.ReadFull(c.dataFile, header)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		timestamp := binary.BigEndian.Uint64(header[0:8])
		keySize := binary.BigEndian.Uint32(header[8:12])
		valueSize := binary.BigEndian.Uint32(header[12:16])
		fmt.Printf("timestamp: %d, keySize: %d, valueSize: %d\n", timestamp, keySize, valueSize)

		// 读取key
		key := make([]byte, keySize)
		if _, err := io.ReadFull(c.dataFile, key); err != nil {
			return err
		}
		// 读取value
		value := make([]byte, valueSize)
		if _, err := io.ReadFull(c.dataFile, value); err != nil {
			return err
		}
		c.items[string(key)] = Item{
			Object:     value,
			Expiration: int64(timestamp),
		}
		index[string(key)] = offset
		offset += 16 + int64(keySize) + int64(valueSize)
	}
	return nil
}
