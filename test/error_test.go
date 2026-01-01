package qiao

import (
	"fmt"
	"testing"
	"time"

	"github.com/chris-liu-zh/qiao"
)

func TestError(t *testing.T) {
	if err := qiao.NewLog().SetViewOut(true).SetDefault(); err != nil {
		t.Fatal(err)
	}
	for k := range 10 {
		qiao.Err("error test", fmt.Errorf("test error%d", k))
	}
	time.Sleep(5 * time.Second)
}
