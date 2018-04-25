package col

import "image"

func (co *Col) Loc() image.Rectangle {
	if co == nil {
		return image.ZR
	}
	return image.Rectangle{co.sp, co.sp.Add(co.size)}
}

func (co *Col) Move(sp image.Point) {
	delta := sp.Sub(co.sp)
	co.Tag.Move(co.Tag.Loc().Min.Add(delta))
	for _, t := range co.List {
		t.Move(t.Loc().Min.Add(delta))
	}
	co.sp = sp
}

func (co *Col) Resize(size image.Point) {
	co.size = size
	notesize(co.Tag)
	pt := image.Pt(co.size.X, co.tdy)
	co.Tag.Resize(pt)

	co.fill()
	for _, k := range co.List {
		notesize(k)
		k.Refresh()
	}
}

type Axis interface {
	Major(image.Point) image.Point
}

func (c *Col) Area() image.Rectangle {
	dy := c.Tag.Loc().Dy()
	return image.Rect(c.sp.X, c.sp.Y+dy, c.sp.X+c.size.X, c.sp.Y+c.size.Y)
}
func (c *Col) Major(pt image.Point) image.Point {
	pt.X = c.sp.X
	pt.Y = clamp(pt.Y, c.Area().Min.Y, c.Area().Max.Y)
	return pt
}

func clamp(v, l, h int) int {
	if v < l {
		return l
	}
	if v > h {
		return h
	}
	return v
}
