package tag

import "image"

type Vis int

const (
	VisNone Vis = 0
	VisTag  Vis = 1
	VisBody Vis = 1 << 1
	VisFull Vis = VisTag | VisBody
)

func Height(faceHeight int) int {
	if faceHeight == 0 {
		faceHeight = DefaultFaceHeight
	}
	if faceHeight != 11 {
		panic(faceHeight)
	}
	return faceHeight * 2
}

func (t *Tag) Move(pt image.Point) {
	t.Win.Move(pt)
	if t.Body == nil {
		return
	}
	pt.Y += t.Win.Loc().Dy()
	t.Body.Move(pt)
}

func (t *Tag) Resize(pt image.Point) {
	ts := t.Config.TagHeight()
	if ts > pt.Y {
		pt.Y = 0
		t.Win.Resize(pt)
		t.Body.Resize(image.ZP)
		t.Vis = VisNone
		return
	}
	t.dirty = true
	if ts*2 > pt.Y {
		pt.Y = ts
		t.Win.Resize(pt)
		t.Body.Resize(image.ZP)
		t.Vis = VisTag
		return
	}
	t.Win.Resize(image.Pt(pt.X, ts))
	t.Body.Resize(image.Pt(pt.X, pt.Y-ts))
	t.Vis = VisFull
}

func (t *Tag) Loc() image.Rectangle {
	r := t.Win.Loc()
	if t.Body != nil {
		r.Max.Y += t.Body.Loc().Dy()
	}
	return r
}

func (t *Tag) Bounds() image.Rectangle {
	return t.Loc()
}
