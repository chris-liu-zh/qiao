package qiao

import (
	"strconv"
	"testing"
	"time"

	"github.com/chris-liu-zh/qiao/cache"
)

func Test_Cache(t *testing.T) {
	kvStore, err := cache.NewKVStore("cache.db")
	if err != nil {
		t.Fatal(err)
		return
	}

	kv, err := cache.New(cache.WithSave(kvStore, 10*time.Second, 500))
	if err != nil {
		t.Fatal(err)
		return
	}
	kv.Flush()
	return
	for i := range 10000 {
		kv.Set("key"+strconv.Itoa(i), i)
		v, _ := kv.Get("key" + strconv.Itoa(i)).Int()
		t.Log(v)
	}
	time.Sleep(15 * time.Second)
}
