package qiao

import (
	"testing"

	"github.com/chris-liu-zh/qiao/tools"
)

func TestNewPtr(t *testing.T) {
	v := 123
	p := tools.NewPtr(v)
	if p == nil {
		t.Errorf("NewPtr returned nil")
	}
	if *p != v {
		t.Errorf("NewPtr returned pointer to %d, expected %d", *p, v)
	}
}
