package qiao

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/chris-liu-zh/qiao/cache"
)

func Test_Cache(t *testing.T) {
	kv, err := cache.New(cache.WithSave("./cache.db", 2*time.Second, 1))
	if err != nil {
		t.Fatal(err)
		return
	}
	kv.Flush()
	kv.Set("key", 123)
	fmt.Println(cache.Increment(kv, "key", 1))

	for i := range 100 {
		kv.Set("key"+strconv.Itoa(i), i)
		time.Sleep(5 * time.Second)
		v, _ := kv.Get("key" + strconv.Itoa(i)).Int()
		t.Log(v)
		kv.Del("key")
	}
}
