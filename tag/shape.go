package tag

import "image"

type Vis int

const (
	VisNone Vis = 0
	VisTag  Vis = 1
	VisBody Vis = 1 << 1
	VisFull Vis = VisTag | VisBody
)

func (t *Tag) Move(pt image.Point) {
	t.sp = pt
	t.Win.Move(pt)
	if t.Body == nil {
		return
	}
	pt.Y += t.Win.Bounds().Dy()
	t.Body.Move(pt)
}

func (t *Tag) Resize(pt image.Point) {
	ts := t.Config.TagHeight()
	if ts > pt.Y {
		pt.Y = 0
		t.size = pt
		t.Win.Resize(pt)
		t.Body.Resize(pt)
		t.Vis = VisNone
		return
	}
	t.dirty = true
	if ts*2 > pt.Y {
		// Theres enough room for the label but the body wouldn't
		// have enough room.
		pt.Y = ts
		t.size = pt
		t.Win.Resize(pt)

		// Coherence: window always under tag
		t.align()

		pt.Y = 0
		t.Body.Resize(pt)
		t.Vis = VisTag
		return
	}
	t.size = pt
	t.Win.Resize(image.Pt(pt.X, ts))
	t.align()
	t.Body.Resize(image.Pt(pt.X, pt.Y-ts))
	t.Vis = VisFull
}

func (t *Tag) align() {
	// Coherence: window always under tag
	r := t.Win.Bounds()
	r.Min.Y = r.Max.Y
	t.Body.Move(r.Min)
}

func (t *Tag) Bounds() image.Rectangle {
	return image.Rectangle{t.sp, t.sp.Add(t.size)}
}
