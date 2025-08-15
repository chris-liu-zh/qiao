package cache

type DirtyOpt int

const (
	DirtyOpInsert DirtyOpt = iota + 1
	DirtyOpUpdate
	DirtyOpDelete
)

// DirtyItem 脏数据项
type DirtyItem struct {
	Key   string
	Value Item
	Opt   DirtyOpt
}

var (
	// DirtyItems 脏数据项
	DirtyItems []DirtyItem
)

func setDirtyItem(key string, value Item, opt DirtyOpt) {
	DirtyItems = append(DirtyItems, DirtyItem{
		Key:   key,
		Value: value,
		Opt:   opt,
	})
}

func ClearDirtyItems() {
	DirtyItems = []DirtyItem{}
}
