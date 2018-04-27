package col

import (
	"fmt"
	"image"

	"github.com/as/font"
	"github.com/as/frame"
	"github.com/as/ui"
	"github.com/as/ui/tag"
)

type Col struct {
	Table2
}

var DefaultConfig = &tag.Config{
	Margin:     image.Pt(15, 0),
	Facer:      font.NewFace,
	FaceHeight: 11,
	Color: [3]frame.Color{
		0: frame.ATag1,
	},
	Ctl: make(chan interface{}, 10),
}

// New creates a new column with the device, font, source point
// and size.
func New(dev ui.Dev, sp, size image.Point, conf *tag.Config) *Col {
	if conf == nil {
		conf = DefaultConfig
	}
	return &Col{
		Table2: NewTable2(dev, sp, size, conf),
	}
}

func (co *Col) Resize(size image.Point) {
	co.size = size
	notesize(co.Tag)
	pt := image.Pt(co.size.X, co.tdy)
	co.Tag.Resize(pt)
	Fill(co)
	for _, k := range co.List {
		notesize(k)
		k.Refresh()
	}
}

func (co *Col) RollDown(id int, dy int) {
}
func (co *Col) MoveWin(id int, y int) {
	if id >= len(co.List) {
		return
	}
	//FLAG
	maxy := co.Loc().Max.Y - co.tdy
	if y >= maxy {
		return
	}
	s := co.detach(id)
	co.Attach(s, y)
}

func (co *Col) Grow(id int, dy int) {
	a, b := id-1, id
	if co.badID(a) || co.badID(b) {
		return
	}
	ra, rb := co.List[a].Loc(), co.List[b].Loc()
	ra.Max.Y -= dy
	if dy := ra.Dy() - co.tdy; dy < 0 {
		co.Grow(a, -dy)
	}
	fmt.Printf("min: %d, dy: %d, min-dy: %d\n", rb.Min.Y, dy, rb.Min.Y-dy)
	co.MoveWin(b, rb.Min.Y-dy)
}

func (co *Col) RollUp(id int, dy int) {
	if id <= 0 || id >= len(co.List) {
		return
	}
	pt := co.Tag.Loc().Min
	pt.Y += co.tdy
	for x := 1; x <= id; x++ {
		pt.Y += co.tdy
		co.List[x].Move(pt)
	}
	Fill(co)
	co.Refresh()
}