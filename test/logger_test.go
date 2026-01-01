package qiao

import (
	"log/slog"
	"testing"
	"time"

	"github.com/chris-liu-zh/qiao"
)

func TestLogger(t *testing.T) {
	if err := qiao.NewLog().SetLevel(slog.LevelDebug).SetViewOut(true).SetDefault(); err != nil {
		t.Fatal(err)
	}
	for range 10 {
		qiao.LogWarn("this is debug log", slog.String("key2", "value2"))
	}
	time.Sleep(5 * time.Second)
}
