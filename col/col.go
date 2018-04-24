package col

import (
	"fmt"
	"image"
	"io"

	"github.com/as/font"
	"github.com/as/frame"
	"github.com/as/shiny/screen"
	"github.com/as/ui"
	"github.com/as/ui/tag"
)

func (c *Col) Dev() ui.Dev                { return c.dev }
func (c *Col) Face() font.Face            { return c.ft }
func (c *Col) ForceSize(size image.Point) { c.size = size }

func NewGridHack(dev ui.Dev, sp, size image.Point, tdy int, ft font.Face) *Col {
	return &Col{
		dev: dev,
		tdy: tdy,
		sp:  sp, size: size, ft: ft,
	}
}

func (c *Col) AttachFill(src Plane, x int) {
	c.attach(src, x)
	c.fill()
}
func (c *Col) DetachFill(src Plane) {
	c.detach(c.ID(src))
	c.fill()
}

type Col struct {
	dev  ui.Dev
	ft   font.Face
	sp   image.Point
	size image.Point
	tdy  int
	Tag  *tag.Tag
	List []Plane
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
	tdy := conf.TagHeight()

	T := tag.New(dev, sp, image.Pt(size.X, tdy), conf)
	T.Win.InsertString("New Delcol Sort	|", 0)

	return &Col{
		dev: dev,
		sp:  sp, size: size,
		ft:   conf.Facer(conf.FaceHeight),
		tdy:  tdy,
		Tag:  T,
		List: make([]Plane, 0),
	}
}

func (co *Col) FindName(name string) *tag.Tag {
	for _, v := range co.List[1:] {
		switch v := v.(type) {
		case *Col:
			t := v.FindName(name)
			if t != nil {
				return t
			}
		case *tag.Tag:
			if v.FileName() == name {
				return v
			}

		}
	}
	return nil
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

func (co *Col) Upload(wind screen.Window) {
	type Uploader interface {
		Upload(screen.Window)
		Dirty() bool
	}
	co.Tag.Upload(wind)
	for _, t := range co.List {
		if t, ok := t.(Uploader); ok {
			//if co.Dirty() {
			t.Upload(wind)
			//	}
		}
	}
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
	co.fill()
	co.Refresh()
}
func (co *Col) IDPoint(pt image.Point) (id int) {
	for id = 0; id < len(co.List); id++ {
		if pt.In(co.List[id].Loc()) {
			break
		}
	}
	return id
}
func (co *Col) ID(w Plane) (id int) {
	for id = 0; id < len(co.List); id++ {
		if w != nil && co.List[id] != nil && w == co.List[id] {
			break
		}
	}
	return id
}

func (co *Col) Close() error {
	co.Tag.Close()
	for _, t := range co.List {
		if t == nil {
			continue
		}
		if t, ok := t.(io.Closer); ok {
			t.Close()
		}
	}
	co.List = nil
	return nil
}
