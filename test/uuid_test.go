package qiao

import (
	"testing"

	"github.com/chris-liu-zh/qiao/tools"
)

func TestUUID(t *testing.T) {
	uu := tools.UUIDV7()
	t.Log(uu.String())
	uu.SetVariant(2)
	t.Log(uu.String())
	t.Log(uu.Variant())
}
