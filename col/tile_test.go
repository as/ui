package col

import (
	"image"
	"testing"

	"github.com/as/ui/tag"
)

func TestDelta(t *testing.T) {
	r := image.Rect(100, 100, 1100, 1100)
	tt := tag.New(etch, r.Min, r.Size(), nil)
	tt2 := tag.New(etch, r.Min, r.Size(), nil)
	c := New(etch, r.Min, r.Size(), nil)
	c.Attach(tt, 300)
	c.Attach(tt2, 400)
	for n, v := range c.List {
		t.Logf("%d: %v %s", n, delta(c, n), v.Loc())
	}
}
