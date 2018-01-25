package win

import (
	"image"
	"image/draw"

	"github.com/as/text"
)

func region5(r0, r1, q0, q1 int) int {
	{
		r0 := int64(r0)
		r1 := int64(r1)
		q0 := int64(q0)
		q1 := int64(q1)
		return text.Region5(r0, r1, q0, q1)
	}
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

func clamp(v, l, h int64) int64 {
	if v < l {
		return l
	}
	if v > h {
		return h
	}
	return v
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
func drawBorder(dst draw.Image, r image.Rectangle, src image.Image, sp image.Point, thick int) {
	draw.Draw(dst, image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+thick), src, sp, draw.Src)
	draw.Draw(dst, image.Rect(r.Min.X, r.Max.Y-thick, r.Max.X, r.Max.Y), src, sp, draw.Src)
	draw.Draw(dst, image.Rect(r.Min.X, r.Min.Y, r.Min.X+thick, r.Max.Y), src, sp, draw.Src)
	draw.Draw(dst, image.Rect(r.Max.X-thick, r.Min.Y, r.Max.X, r.Max.Y), src, sp, draw.Src)
}
