package qiao

import (
	"errors"
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
	if err := a(); err != nil {
		return
	}
	time.Sleep(5 * time.Second)
}

func a() error {
	if err := b(); err != nil {
		return qiao.Err("error test", err)
	}
	return nil
}

func b() error {
	if err := c(); err != nil {
		return qiao.Err("error test", err)
	}
	return nil
}

func c() error {
	return qiao.Err("error test", errors.New("test error"), qiao.SetOther(map[string]string{"key": "value"}))
}
