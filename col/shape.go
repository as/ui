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
	co.fill()
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
