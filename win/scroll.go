package win

import (
	"image"
	"image/draw"

	"github.com/as/frame"
	"github.com/as/text"
)

const minSbWidth = 10

type Drawer interface {
	Draw(dst draw.Image, r image.Rectangle, src image.Image, sp image.Point, op draw.Op)
}

type ScrollBar struct {
	r       image.Rectangle
	bar     image.Rectangle
	lastbar image.Rectangle
	Fg, Bg  image.Image
}

func (s *ScrollBar) drawsb(dst draw.Image, d Drawer) bool {
	if s.r == image.ZR {
		return false
	}
	draw0 := draw.Draw
	if d != nil {
		draw0 = d.Draw
	}

	r0, r1, q0, q1 := s.bar.Min.Y, s.bar.Max.Y, s.lastbar.Min.Y, s.lastbar.Max.Y
	//s.lastbar = s.bar
	r := s.bar
	switch region5(r0, r1, q0, q1) {
	case -2, 2, 0: //w.Frame.RGBA()
		draw0(dst, image.Rect(r.Min.X, q0, r.Max.X, q1), s.Bg, image.ZP, draw.Src)
		draw0(dst, image.Rect(r.Min.X, r0, r.Max.X, r1), s.Fg, image.ZP, draw.Src)
	case -1:
		draw0(dst, image.Rect(r.Min.X, r1, r.Max.X, q1), s.Bg, image.ZP, draw.Src)
		draw0(dst, image.Rect(r.Min.X, r0, r.Max.X, q0), s.Fg, image.ZP, draw.Src)
	case 1:
		draw0(dst, image.Rect(r.Min.X, q0, r.Max.X, r0), s.Bg, image.ZP, draw.Src)
		draw0(dst, image.Rect(r.Min.X, q1, r.Max.X, r1), s.Fg, image.ZP, draw.Src)
	}
	return true
}
func (s ScrollBar) refreshsb(dst draw.Image, d Drawer) {
	draw0 := draw.Draw
	if d != nil {
		draw0 = d.Draw
	}
	draw0(dst, s.r, s.Bg, image.ZP, draw.Src)
	draw0(dst, s.bar, s.Fg, image.ZP, draw.Src)
}

func NewScrollBar(r image.Rectangle, fg, bg image.Image) (sb ScrollBar) {
	sb.r = r
	sb.Fg = fg
	sb.Bg = bg
	if bg == nil {
		sb.Bg = frame.ATag0.Back
	}
	if fg == nil {
		sb.Fg = LtGray
	}
	return sb
}

func (s *ScrollBar) update(advance, cover float64) bool {
	r := s.r
	if r == image.ZR {
		return false
	}

	r.Min.Y += int(float64(r.Max.Y) * advance)
	r.Max.Y = int(float64(r.Max.Y) * cover)
	if have := r.Max.Y - r.Min.Y; have < 3 {
		r.Max.Y = r.Min.Y + 3
	}

	r.Min.Y = clamp32(r.Min.Y, s.r.Min.Y, s.r.Max.Y)
	r.Max.Y = clamp32(r.Max.Y, s.r.Min.Y, s.r.Max.Y)

	//	if s.bar == r{
	//		return false
	//	}
	s.lastbar = s.bar
	s.bar = r
	return true
}

func (w *Win) updatesb() {
	if w.ScrollBar.r == image.ZR {
		return
	}
	rat0 := float64(w.org) / float64(w.Len())               // % scrolled
	rat1 := float64(w.org+w.Frame.Len()) / float64(w.Len()) // % covered by screen
	w.dirty = w.ScrollBar.update(rat0, rat1) || w.dirty
}

func (w *Win) scrollinit(pad image.Point) {
	if w.Frame == nil {
		return
	}
	sr := w.Frame.RGBA().Bounds()
	if pad.X > minSbWidth+2 {
		sr.Max.X = minSbWidth
	}
	w.ScrollBar = NewScrollBar(sr, nil, nil)
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
		rat := float64(pt.Y) / float64(w.ScrollBar.r.Dy())
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
	w.dirty = w.ScrollBar.drawsb(w.Frame.RGBA(), w.Frame) || w.dirty
}
func (w *Win) refreshsb() {
	w.ScrollBar.refreshsb(w.Frame.RGBA(), w.Frame)
	w.dirty = true
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
