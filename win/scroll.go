package win

import (
	"github.com/as/frame"
	"github.com/as/text"
	"image"
	"image/draw"
)

const minSbWidth = 10

type ScrollBar struct {
	bar     image.Rectangle
	Scrollr image.Rectangle
}

func (w *Win) scrollinit(pad image.Point) {
	w.Scrollr = image.ZR
	if pad.X > minSbWidth+2 {
		sr := w.Frame.RGBA().Bounds()
		sr.Max.X = minSbWidth
		w.Scrollr = sr
	}
	w.Frame.Draw(w.Frame.RGBA(), w.Scrollr, frame.ATag0.Back, image.ZP, draw.Src)
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
		if mul == 0{
			mul++
		}
		dx := w.IndexOf(image.Pt(r.Min.X, r.Min.Y+dl*w.Font.Dy()))*mul
		org += dx
		w.SetOrigin(org, true)
	}
	w.updatesb()
	w.drawsb()
	w.dirty = true
}

func region3(r, q0, q1 int) int {
	return text.Region3(int64(r), int64(q0), int64(q1))
}
func (w *Win) Clicksb(pt image.Point, dir int) {
	n := 0
	for region3(pt.Y, w.bar.Min.Y-3, w.bar.Min.Y+3) != 0 {
		if n == 4 {
			break
		}
		w.clicksb(pt, dir)
		n++
	}
	w.drawsb()
	w.dirty = true
}
func (w *Win) clicksb(pt image.Point, dir int) {
	var (
		rat float64
	)
	fl := float64(w.Frame.Len())
	n := w.org
	barY0 := float64(w.bar.Min.Y)
	barY1 := float64(w.bar.Max.Y)
	ptY := float64(pt.Y)
	switch dir {
	case -1:
		rat = barY1 / ptY
		delta := int64(fl * rat)
		n -= delta
	case 0:
		rat = (ptY - barY0) / (barY1 - barY0)
		delta := int64(fl * rat)
		n += delta
	case 1:
		rat = (barY1 / ptY)
		delta := int64(fl * rat)
		n += delta
	}
	w.SetOrigin(n, false)
	w.updatesb()
}

func (w *Win) realsbr(r image.Rectangle) image.Rectangle {
	return r.Add(w.Sp).Add(image.Pt(0, w.pad.Y))
}

func (w *Win) drawsb() {
	w.Frame.Draw(w.Frame.RGBA(), w.Scrollr, frame.ATag0.Back, image.ZP, draw.Src)
	w.Frame.Draw(w.Frame.RGBA(), w.bar, LtGray, image.ZP, draw.Src)
}

func (w *Win) updatesb() {
	r := w.Scrollr
	dy := float64(w.Frame.Bounds().Dy() - w.pad.Y)
	rat0 := float64(w.org) / float64(w.Len()) // % scrolled
	r.Min.Y = +int(dy * rat0)

	rat1 := float64(w.org+w.Frame.Len()) / float64(w.Len()) // % covered by screen
	r.Max.Y = int(dy * rat1)
	if r.Max.Y-r.Min.Y < 3 {
		r.Max.Y = r.Min.Y + 3
	}
	w.dirty = true
	w.bar = r
}
