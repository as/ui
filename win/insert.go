package win

import (
	"io"

	"github.com/as/text"
)

// Insert inserts the bytes in p at position q0. When q0
// is zero, Insert prepends the bytes in p to the underlying
// buffer
func (w *Win) Insert(p []byte, q0 int64) (n int) {
	if len(p) == 0 {
		return 0
	}

	// If at least one point in the region overlaps the
	// frame's visible area then we alter the frame. Otherwise
	// there's no point in moving text down, it's just annoying.

	switch q1 := q0 + int64(len(p)); text.Region5(q0, q1, w.org-1, w.org+w.Frame.Len()+1) {
	case -2:
		w.org += q1 - q0
	case -1:
		// Insertion to the left
		w.Frame.Insert(p[q1-w.org:], 0)
		w.org += w.org - q0
	case 1:
		w.Frame.Insert(p, q0-w.org)
	case 0:
		if q0 < w.org {
			p0 := w.org - q0
			w.Frame.Insert(p[p0:], 0)
			w.org += w.org - q0
		} else {
			w.Frame.Insert(p, q0-w.org)
		}
	}
	if w.Editor == nil {
		panic("nil editor")
	}
	n = w.Editor.Insert(p, q0)
	return n
}

func (w *Win) WriteAt(p []byte, at int64) (n int, err error) {
	n, err = w.Editor.(io.WriterAt).WriteAt(p, at)
	q0, q1 := at, at+int64(len(p))

	switch text.Region5(q0, q1, w.org-1, w.org+w.Frame.Len()+1) {
	case -2:
		// Logically adjust origin to the left (up)
		w.org -= q1 - q0
	case -1:
		// Remove the visible text and adjust left
		w.Frame.Delete(0, q1-w.org)
		w.org = q0
		w.Fill()
	case 0:
		p0 := clamp(q0-w.org, 0, w.Frame.Len())
		p1 := clamp(q1-w.org, 0, w.Frame.Len())
		w.Frame.Delete(p0, p1)
		w.Fill()
	case 1:
		w.Frame.Delete(q0-w.org, w.Frame.Len())
		w.Fill()
	case 2:
	}
	return
}
