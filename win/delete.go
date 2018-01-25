package win

import (
	"github.com/as/text"
)

// Delete deletes the range [q0:q1] inclusive. If there
// is nothing to delete, it returns 0.
func (w *Win) Delete(q0, q1 int64) (n int) {
	if w.Len() == 0 {
		return 0
	}
	if q0 > q1 {
		q0, q1 = q1, q0
	}
	if q1 > w.Len() {
		q1 = w.Len()
	}
	w.Editor.Delete(q0, q1)

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
	return int(q1 - q0)
}
