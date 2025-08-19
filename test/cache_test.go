package qiao

import (
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

	kv, err := cache.New(cache.WithSave(kvStore, 2*time.Second, 100))
	if err != nil {
		t.Fatal(err)
		return
	}
	kv.Set("key", "value")

}
