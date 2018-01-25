package win

import (
	"image"

	"github.com/as/text"
)

const (
	// Extra lines to scroll down to comfortably display the result of a look operation
	JumpScrollMargin = -3
)

// Jump scrolls the active selection into view. An optional mouseFunc
// is given the transfer coordinates to move the mouse cursor under
// the selection.
func (w *Win) Jump(mouseFunc func(image.Point)) {
	q0, q1 := w.Dot()
	if text.Region5(q0, q1, w.Origin(), w.Origin()+w.Frame.Len()) != 0 {
		w.SetOrigin(q0, true)
		w.Scroll(JumpScrollMargin)
	}
	if mouseFunc != nil {
		jmp := w.PointOf(q0 - w.org)
		mouseFunc(w.Bounds().Min.Add(jmp))
	}
}
