package qiao

import "log/slog"

func init() {
	if err := NewLog().SetLog(); err != nil {
		slog.Error("init logger error", "error", err)
	}
}
