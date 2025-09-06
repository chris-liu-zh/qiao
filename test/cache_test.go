package qiao

import (
	"fmt"
	"testing"
	"time"

	"github.com/chris-liu-zh/qiao/cache"
)

func Test_Cache(t *testing.T) {
	fmt.Println("start1111")
	kv, err := cache.New()
	if err != nil {
		t.Fatal(err)
		return
	}
	start := time.Now()
	for i := range 10000 {
		key := fmt.Sprintf("key%05d", i)
		val := fmt.Sprintf("value%05d", i)
		kv.Set(key, val, 60*time.Second)
	}

	fmt.Println("set cost:", time.Since(start).String())
	kv.Count()
	fmt.Println("end1111")
	time.Sleep(5 * time.Second)
}
