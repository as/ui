package win

import (
	"github.com/as/text"
)

// Select selects the range [q0:q1] inclusive
func (w *Win) Select(q0, q1 int64) {
	if q0 > q1 {
		q0, q1 = q1, q0
	}
	q00, q11 := w.Dot()
	w.Editor.Select(q0, q1)
	reg := text.Region3(q0, w.org-1, w.org+w.Frame.Len())
	if q00 == q0 && q11 == q1 {
		//return
	}
	p0, p1 := q0-w.org, q1-w.org
	w.Frame.Select(p0, p1)
	if q0 == q1 && reg != 0 {
		//w.Untick()	// TODO(as): win.exe cursor disappeared when this was uncommented
	}
}
