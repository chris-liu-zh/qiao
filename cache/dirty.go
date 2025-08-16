package cache

import "time"

type DirtyOpt int

const (
	DirtyOpPut DirtyOpt = iota + 1
	DirtyOpDel
)

type DirtyItem struct {
	Key    string
	Value  []byte
	Expire int64
	Opt    DirtyOpt
}

func (c *cache) makeDirty() {
	c.dirtyItems = make(map[int64]DirtyItem)
	c.sortDirtyKeys = make([]int64, 0)
}

func (c *cache) dirtyPut(key string, value []byte, expire int64, opt DirtyOpt) {
	id := time.Now().UnixNano()
	if opt == DirtyOpDel {
		c.dirtyItems[id] = DirtyItem{
			Key: key,
			Opt: opt,
		}
		c.setDirtyKey(id)
		return
	}
	if expire <= time.Now().Unix() {
		return
	}

	c.dirtyItems[id] = DirtyItem{
		Key:    key,
		Value:  value,
		Expire: expire,
		Opt:    opt,
	}
	c.setDirtyKey(id)
}

func (c *cache) flushDirty() {
	c.dirtyItems = make(map[int64]DirtyItem)
	c.sortDirtyKeys = make([]int64, 0)
	c.DirtyTotal = 0
}

func (c *cache) setDirtyKey(id int64) {
	c.sortDirtyKeys = append(c.sortDirtyKeys, id)
	c.DirtyTotal++
}
