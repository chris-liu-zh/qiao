package qiao

import (
	"fmt"
	"testing"
	"time"

	"github.com/chris-liu-zh/qiao/DB"
	"github.com/chris-liu-zh/qiao/cache"
)

func Test_Cache(t *testing.T) {
	if err := initialize(); err != nil {
		t.Fatalf("%v", err)
		return
	}
	var products []Ptype
	if err := DB.QiaoDB().GetList(&products); err != nil {
		t.Fatalf("%v", err)
		return
	}
	fmt.Println("start")
	kvStore, err := cache.NewKVStore("cache.db")
	if err != nil {
		t.Fatal(err)
		return
	}

	kv, err := cache.New(cache.WithSave(kvStore, 1*time.Second, 1000))
	if err != nil {
		t.Fatal(err)
		return
	}
	kv.Flush()
	for _, product := range products {
		kv.Set(product.Typeid, product)
		// kv.Del(product.Typeid)
	}
	time.Sleep(10 * time.Second)

}
