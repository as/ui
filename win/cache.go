package win

import (
	"image"

	"github.com/as/shrew"
)

func (w *Win) Buffer() {
	w.cacher.buffer = true
}
func (w *Win) Unbuffer() {
	w.cacher.buffer = false
	w.Flush()
}

type cacher struct {
	buffer bool
	r      []image.Rectangle
	shrew.Bitmap
}

func (c *cacher) Flush(r ...image.Rectangle) error {
	if c.buffer {
		c.r = append(c.r, r...)
	} else {
		c.Bitmap.Flush(append(c.r, r...)...)
		c.r = c.r[:0]
	}
	return nil
}
