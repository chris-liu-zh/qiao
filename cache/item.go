package cache

import "time"

type ObjectType interface {
	string | int | int64 | uint | uint64 | float32 | float64 | bool | []byte
}

// Expired 如果项目已过期，则返回 true。
func (item Item) Expired() bool {
	if item.Expiration <= 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

// SetInvalid 将项目设置为无效状态，并设置错误信息。
func (item Item) setInvalid(err error) Item {
	item.invalid = true
	item.err = err
	return item
}

func (item Item) String() (string, error) {
	return result[string](item)
}

func (item Item) Int() (int, error) {
	return result[int](item)
}

func (item Item) Int64() (int64, error) {
	return result[int64](item)
}

func (item Item) Uint() (uint, error) {
	return result[uint](item)
}

func (item Item) Uint64() (uint64, error) {
	return result[uint64](item)
}

func (item Item) Float32() (float32, error) {
	return result[float32](item)
}

func (item Item) Float64() (float64, error) {
	return result[float64](item)
}

func (item Item) Bool() (bool, error) {
	return result[bool](item)
}

func (item Item) Bytes() ([]byte, error) {
	return result[[]byte](item)
}

func (item Item) Scan(val any) error {
	return gobDecode(item.Object, val)
}

func result[T ObjectType](item Item) (T, error) {
	var val T
	if item.err != nil {
		return val, item.err
	}
	err := item.Scan(&val)
	return val, err
}
