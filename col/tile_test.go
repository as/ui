package col

import (
	"image"
	"testing"

	"github.com/as/ui/tag"
)

func TestDelta(t *testing.T) {
	tt := tag.New(etch, nil)
	tt2 := tag.New(etch, nil)
	c := New(etch, nil)
	r := image.Rect(100, 100, 1100, 1100)
	c.Move(r.Min)
	c.Resize(r.Size())
	c.Attach(tt, 300)
	c.Attach(tt2, 400)
	for n, v := range c.List {
		t.Logf("%d: %v %s", n, delta(c, n), v.Loc())
	}
}
