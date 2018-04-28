// Package scroll implements a vertical scroll bar
package scroll

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/as/text"
)

const (
	Min = 10
)

var (
	Mauve  = image.NewUniform(color.RGBA{216, 216, 232, 255})
	LtGray = image.NewUniform(color.RGBA{128 + 32 + 16, 128 + 32 + 16, 128 + 32 + 16 + 16, 255})
)

var (
	DefaultColors = [...]image.Image{LtGray, Mauve}
)

type Drawer interface {
	Draw(dst draw.Image, r image.Rectangle, src image.Image, sp image.Point, op draw.Op)
}

// Bar is a scrollbar, currently vertical only
type Bar struct {
	r       image.Rectangle
	bar     image.Rectangle
	lastbar image.Rectangle
	fg, bg  image.Image
}

// New initializes using r as the bounds using fg and bg as the
// foreground and background colors. Default colors are used
// if fg or bg are nil.
func New(r image.Rectangle, fg, bg image.Image) (b Bar) {
	b.r = r
	b.fg = fg
	b.bg = bg
	if fg == nil {
		b.fg = DefaultColors[0]
	}
	if bg == nil {
		b.bg = DefaultColors[1]
	}
	return b
}

// Put updates the delta and coverage values for the bar. The delta
// is the ratio representing how far down the bar is currently scrolled relative
// to the entire document. The cover is the ratio of the document that is currently
// viewable by the client. Both values are ranges between [0.0, 1.0]
//
// The delta 1.0 is valid, and means that the document's contents are beyond the
// scroll bars representative client area.
func (s *Bar) Put(delta, cover float64) bool {
	if s.r == image.ZR {
		return false
	}
	r := s.r
	r.Min.Y += int(float64(r.Max.Y) * delta)
	r.Max.Y = int(float64(r.Max.Y) * cover)
	if have := r.Max.Y - r.Min.Y; have < 3 {
		r.Max.Y = r.Min.Y + 3
	}

	r.Min.Y = clamp(r.Min.Y, s.r.Min.Y, s.r.Max.Y)
	r.Max.Y = clamp(r.Max.Y, s.r.Min.Y, s.r.Max.Y)

	//	if s.bar == r{
	//		return false
	//	}
	s.lastbar = s.bar
	s.bar = r
	return true
}

// Delta returns the ratio representing how far down bar would be scrolled
// it pt was its starting point.
func (s Bar) Delta(pt image.Point) float64 {
	return float64(pt.Y) / float64(s.r.Dy())
}

// Cover returns the coverage ratio of the current bar if its representative window
// had the dimensions of the given size. Content is assumed to be uniform, which
// may not be the case with line-wrapped text.
func (s Bar) Cover(size image.Point) float64 {
	return float64(s.bar.Dy()) / float64(size.Y)
}

// Update draws the modified regions of the bar on dst using an
// optional drawer. A nil drawer retults in the standard draw.Draw
// call.
//
func (s *Bar) Update(dst draw.Image, d Drawer) bool {
	if s.r == image.ZR {
		return false
	}
	draw0 := draw.Draw
	if d != nil {
		draw0 = d.Draw
	}

	r0, r1, q0, q1 := s.bar.Min.Y, s.bar.Max.Y, s.lastbar.Min.Y, s.lastbar.Max.Y
	r := s.bar
	switch region5(r0, r1, q0, q1) {
	case -2, 2, 0:
		draw0(dst, image.Rect(r.Min.X, q0, r.Max.X, q1), s.bg, image.ZP, draw.Src)
		draw0(dst, image.Rect(r.Min.X, r0, r.Max.X, r1), s.fg, image.ZP, draw.Src)
	case -1:
		draw0(dst, image.Rect(r.Min.X, r1, r.Max.X, q1), s.bg, image.ZP, draw.Src)
		draw0(dst, image.Rect(r.Min.X, r0, r.Max.X, q0), s.fg, image.ZP, draw.Src)
	case 1:
		draw0(dst, image.Rect(r.Min.X, q0, r.Max.X, r0), s.bg, image.ZP, draw.Src)
		draw0(dst, image.Rect(r.Min.X, q1, r.Max.X, r1), s.fg, image.ZP, draw.Src)
	}
	return true
}

// Refresh draws the entire scrollbar on dst using an optional drawer
func (s Bar) Refresh(dst draw.Image, d Drawer) {
	if s.r == image.ZR {
		return
	}
	draw0 := draw.Draw
	if d != nil {
		draw0 = d.Draw
	}
	draw0(dst, s.r, s.bg, image.ZP, draw.Src)
	draw0(dst, s.bar, s.fg, image.ZP, draw.Src)
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

func region5(r0, r1, q0, q1 int) int {
	{
		r0 := int64(r0)
		r1 := int64(r1)
		q0 := int64(q0)
		q1 := int64(q1)
		return text.Region5(r0, r1, q0, q1)
	}
}
