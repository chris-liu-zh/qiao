package qiao

import (
	"testing"
	"time"

	"github.com/chris-liu-zh/qiao/cache"
)

func Test_Cache(t *testing.T) {

	// if err := initialize(); err != nil {
	// 	t.Fatalf("%v", err)
	// 	return
	// }
	// var products []Ptype
	// if err := DB.QiaoDB().GetList(&products); err != nil {
	// 	t.Fatalf("%v", err)
	// 	return
	// }
	// var p Ptype
	// count, err := DB.QiaoDB().Count(&p, "typeid")
	// if err != nil {
	// 	t.Fatalf("%v", err)
	// 	return
	// }
	// fmt.Println(count)

	kvStore, err := cache.NewKVStore("cache.db")
	if err != nil {
		t.Fatal(err)
		return
	}

	kv, err := cache.New(cache.WithSave(kvStore, 2*time.Second, 1))
	if err != nil {
		t.Fatal(err)
		return
	}
	// kv.Flush()
	kv.Set("test", "test2")

	var v string
	kv.Get("test").Scan(&v)
	t.Log(v)
	// for _, product := range products {
	// 	// kv.Set(product.Typeid, product)
	// 	// var ptype Ptype
	// 	// kv.Get(product.Typeid).Scan(&ptype)
	// 	// t.Log(ptype.FullName)
	// }
	time.Sleep(4 * time.Second)

}
