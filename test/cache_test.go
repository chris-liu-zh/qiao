package qiao

import (
	"fmt"
	"testing"
	"time"

	"github.com/chris-liu-zh/qiao/cache"
)

func Test_Cache(t *testing.T) {
	fmt.Println("start1111")
	kv, err := cache.New(cache.WithSave("cache.db", 2, 1000))
	if err != nil {
		t.Fatal(err)
		return
	}
	for i := range 100000 {
		key := fmt.Sprintf("key%05d", i)
		val := fmt.Sprintf("value%05d", i)
		kv.Set(key, val, 60*time.Second)
	}
	fmt.Println("end1111")
	time.Sleep(5 * time.Second)
}
