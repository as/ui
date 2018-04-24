package col

import (
	"image"
	"log"
)

type Tile interface {
	Delta(int) image.Point
	Kid(n int) Plane
	Len() int
}

func (co *Col) Attach(src Plane, y int) {
	if y < co.sp.Y || y > co.sp.Y+co.size.Y {
		return // TODO(as): panic
		log.Printf("y out of bounds %d -> %s", y, src.Loc())
		panic("suicide")
	}
	pt := image.Pt(co.sp.X, co.sp.Y+co.Tag.Loc().Dy())
	if len(co.List) == 0 {
		src.Move(pt)
		co.attach(src, 0)
		co.fill()
		return
	}
	pt.Y = y
	src.Move(pt)
	did := co.IDPoint(pt)
	co.attach(src, did)
	co.fill()
}

// attach inserts w in position id, shifting the original forwards
func (co *Col) attach(w Plane, id int) {
	if id == len(co.List) {
		co.List = append(co.List, w)
		return
	}
	log.Printf("id=%v len=%v\n", id, len(co.List))
	co.List = append(co.List[:id], append([]Plane{w}, co.List[id:]...)...)
}

func (co *Col) Detach(id int) Plane {
	return co.detach(id)
}

// detach (logical)
func (co *Col) detach(id int) Plane {
	if id < 0 || id >= len(co.List) {
		return nil
	}
	w := co.List[id]
	copy(co.List[id:], co.List[id+1:])
	co.List = co.List[:len(co.List)-1]
	return w
}

func (co *Col) Fill() {
	co.fill()
}

func (co *Col) fill() {
	fill(co)
	pt := image.Pt(co.size.X, co.size.Y)
	co.Tag.Resize(pt)
}

func fill(t Tile) {
	if t.Len() == 0 {
		return
	}
	for n := 0; n != t.Len(); n++ {
		pt := t.Delta(n)
		if pt == image.ZP {
			return // TODO(as): panic here
			panic("zp")
		}
		k := t.Kid(n)
		k.Resize(pt)
	}
}

func (c *Col) Delta(n int) image.Point {
	y0 := c.Tag.Loc().Max.Y
	y1 := c.Loc().Max.Y

	if n+1 != len(c.List) {
		y1 = c.List[n+1].Loc().Min.Y
	}
	return identity(c.size.X, y1-y0)
}

func identity(x, y int) image.Point {
	if x == 0 || y == 0 {
		return image.ZP
	}
	return image.Pt(x, y)
}
