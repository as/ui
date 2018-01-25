package win

import (
	"image"
	"image/draw"

	"github.com/as/frame"
)

const minSbWidth = 9

type ScrollBar struct {
	bar     image.Rectangle
	Scrollr image.Rectangle
	lastbar image.Rectangle
}

func (w *Win) Scroll(dl int) {
	if dl == 0 {
		return
	}
	org := w.org
	if dl < 0 {
		org = w.BackNL(org, -dl)
		w.SetOrigin(org, true)
	} else {
		if org+w.Frame.Nchars == w.Len() {
			return
		}
		r := w.Frame.Bounds()
		mul := int64(dl / w.Frame.Line())
		if mul == 0 {
			mul++
		}
		dx := w.IndexOf(image.Pt(r.Min.X, r.Min.Y+dl*frame.Dy(w.Font))) * mul
		org += dx
		w.SetOrigin(org, false)
	}
	w.updatesb()
	w.drawsb()

}

func (w *Win) refreshsb() {
	w.c.Draw(w.c, w.Scrollr, frame.ATag0.Back, image.ZP, draw.Src)
	w.c.Draw(w.c, w.bar, LtGray, image.ZP, draw.Src)
	w.c.Flush(w.Scrollr, w.bar)
}

func (w *Win) updatesb() {
	r := w.Scrollr
	if r == image.ZR {
		return
	}
	rat0 := float64(w.org) / float64(w.Len()) // % scrolled
	r.Min.Y += int(float64(r.Max.Y) * rat0)

	rat1 := float64(w.org+w.Frame.Len()) / float64(w.Len()) // % covered by screen
	r.Max.Y = int(float64(r.Max.Y) * rat1)                  //int(dy * rat1)
	if have := r.Max.Y - r.Min.Y; have < 3 {
		r.Max.Y = r.Min.Y + 3
	}

	r.Min.Y = clamp32(r.Min.Y, w.Scrollr.Min.Y, w.Scrollr.Max.Y)
	r.Max.Y = clamp32(r.Max.Y, w.Scrollr.Min.Y, w.Scrollr.Max.Y)
	w.lastbar = w.bar
	w.bar = r
}
func (w *Win) drawsb() {
	if w.Scrollr == image.ZR {
		return
	}
	if w.bar == w.lastbar {
		return
	}
	r0, r1, q0, q1 := w.bar.Min.Y, w.bar.Max.Y, w.lastbar.Min.Y, w.lastbar.Max.Y
	w.lastbar = w.bar
	r := w.bar

	drawfn := func(r image.Rectangle, c image.Image) {
		r.Min.X = w.Scrollr.Min.X
		r.Max.X = w.Scrollr.Max.X
		if r.Max.Y == 0 {
			r.Max.Y = w.Scrollr.Max.Y
		}
		w.c.Draw(w.c, r, c, image.ZP, draw.Src)
		w.c.Flush(r)
	}
	switch region5(r0, r1, q0, q1) {
	case -2, 2, 0:
		drawfn(image.Rect(r.Min.X, q0, r.Max.X, q1), frame.ATag0.Back)
		drawfn(image.Rect(r.Min.X, r0, r.Max.X, r1), LtGray)
	case -1:
		drawfn(image.Rect(r.Min.X, r1, r.Max.X, q1), frame.ATag0.Back)
		drawfn(image.Rect(r.Min.X, r0, r.Max.X, q0), LtGray)
	case 1:
		drawfn(image.Rect(r.Min.X, q0, r.Max.X, r0), frame.ATag0.Back)
		drawfn(image.Rect(r.Min.X, q1, r.Max.X, r1), LtGray)
	}
}
func (w *Win) clicksb(pt image.Point, dir int) {
	var (
		rat float64
	)
	fl := float64(w.Frame.Len())
	n := w.org
	barY1 := float64(w.bar.Max.Y)
	ptY := float64(pt.Y)
	switch dir {
	case -1:
		rat = barY1 / ptY
		delta := int64(fl * rat)
		n -= delta
	case 0:
		rat := float64(pt.Y) / float64(w.Scrollr.Dy())
		w.SetOrigin(int64(float64(w.Len())*rat), false)
		w.updatesb()
		return
	case 1:
		rat = (barY1 / ptY)
		delta := int64(fl * rat)
		n += delta
	}
	w.SetOrigin(n, false)
	w.updatesb()
}
func (w *Win) scrollinit(r image.Rectangle) {
	s := w.c.Bounds()
	r.Max, r.Min = r.Min, s.Min
	r.Max.Y = s.Max.Y
	r.Max.X = r.Min.X + minSbWidth
	w.Scrollr = r
	w.updatesb()
	w.drawsb()
}
