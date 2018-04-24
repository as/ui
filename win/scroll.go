package win

import (
	"image"
	"image/draw"

	"github.com/as/frame"
	"github.com/as/text"
)

const minSbWidth = 10

type ScrollBar struct {
	bar     image.Rectangle
	Scrollr image.Rectangle
	lastbar image.Rectangle
}

func (w *Win) scrollinit(pad image.Point) {
	w.Scrollr = image.ZR
	if pad.X > minSbWidth+2 {
		sr := w.Frame.RGBA().Bounds()
		sr.Max.X = minSbWidth
		w.Scrollr = sr
	}
	w.updatesb()
	w.refreshsb()
}

func (w *Win) Scroll(dl int) {
	if dl == 0 {
		return
	}
	org := w.org
	if dl < 0 {
		org = w.backNL(org, -dl)
		w.SetOrigin(org, true)
	} else {
		// TODO(as): Forward scrolling will be broken in non-graphical mode
		// Needs to be fixed here
		if !w.graphical() {
			return
		}
		if org+w.Frame.Nchars == w.Len() {
			return
		}
		r := w.Frame.Bounds()
		nline := w.Frame.Line()
		if nline == 0 {
			nline = 1
		}
		mul := int64(dl / nline)
		if mul == 0 {
			mul++
		}
		dx := w.IndexOf(image.Pt(r.Min.X, r.Min.Y+dl*w.Face.Dy())) * mul
		org += dx
		w.SetOrigin(org, true)
	}
	w.updatesb()
	w.drawsb()
	w.dirty = true
}

func (w *Win) Clicksb(pt image.Point, dir int) {
	w.clicksb(pt, dir)
	w.drawsb()
	w.dirty = true
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

func region5(r0, r1, q0, q1 int) int {
	{
		r0 := int64(r0)
		r1 := int64(r1)
		q0 := int64(q0)
		q1 := int64(q1)
		return text.Region5(r0, r1, q0, q1)
	}
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
	w.dirty = true
	switch region5(r0, r1, q0, q1) {
	case -2, 2, 0:
		w.Frame.Draw(w.Frame.RGBA(), image.Rect(r.Min.X, q0, r.Max.X, q1), frame.ATag0.Back, image.ZP, draw.Src)
		w.Frame.Draw(w.Frame.RGBA(), image.Rect(r.Min.X, r0, r.Max.X, r1), LtGray, image.ZP, draw.Src)
	case -1:
		w.Frame.Draw(w.Frame.RGBA(), image.Rect(r.Min.X, r1, r.Max.X, q1), frame.ATag0.Back, image.ZP, draw.Src)
		w.Frame.Draw(w.Frame.RGBA(), image.Rect(r.Min.X, r0, r.Max.X, q0), LtGray, image.ZP, draw.Src)
	case 1:
		w.Frame.Draw(w.Frame.RGBA(), image.Rect(r.Min.X, q0, r.Max.X, r0), frame.ATag0.Back, image.ZP, draw.Src)
		w.Frame.Draw(w.Frame.RGBA(), image.Rect(r.Min.X, q1, r.Max.X, r1), LtGray, image.ZP, draw.Src)
		//	case 0:
		//		col := frame.ATag0.Back // for a shrinking bar
		//		if r0 < q0 {            // bar grows larger
		//			col = LtGray
		//		}
		//		w.Frame.Draw(w.Frame.RGBA(), image.Rect(r.Min.X, r0, r.Max.X, q0), col, image.ZP, draw.Src)
		//		w.Frame.Draw(w.Frame.RGBA(), image.Rect(r.Min.X, q1, r.Max.X, r1), col, image.ZP, draw.Src)
	}
}
func (w *Win) refreshsb() {
	w.Frame.Draw(w.Frame.RGBA(), w.Scrollr, frame.ATag0.Back, image.ZP, draw.Src)
	w.Frame.Draw(w.Frame.RGBA(), w.bar, LtGray, image.ZP, draw.Src)
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
	w.dirty = true
	r.Min.Y = clamp32(r.Min.Y, w.Scrollr.Min.Y, w.Scrollr.Max.Y)
	r.Max.Y = clamp32(r.Max.Y, w.Scrollr.Min.Y, w.Scrollr.Max.Y)
	w.lastbar = w.bar
	w.bar = r
}
func clamp32(v, l, h int) int {
	if v < l {
		return l
	}
	if v > h {
		return h
	}
	return v
}
