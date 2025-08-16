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

var (
	DirtyItems = make(map[int64]DirtyItem)
	sortKeys   = make([]int64, 0)
)

func dirtyPut(key string, value []byte, expire int64, opt DirtyOpt) {
	id := time.Now().UnixNano()
	if opt == DirtyOpDel {
		DirtyItems[id] = DirtyItem{
			Key: key,
			Opt: opt,
		}
		sortKeys = append(sortKeys, id)
		return
	}
	if expire <= time.Now().Unix() {
		return
	}
	sortKeys = append(sortKeys, id)
	DirtyItems[id] = DirtyItem{
		Key:    key,
		Value:  value,
		Expire: expire,
		Opt:    opt,
	}

}

func FlushDirtyItems() {
	DirtyItems = make(map[int64]DirtyItem)
	sortKeys = make([]int64, 0)
}
