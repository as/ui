package tag

import (
	"github.com/as/font"
)

type facer interface {
	Face() font.Face
	SetFont(font.Face)
}

// Height returns the recommended minimum pixel height for a tag label
// given the face height in pixels.
func Height(facePix int) int {
	if facePix == 0 {
		facePix = DefaultConfig.FaceHeight
	}
	return facePix + facePix/2 + facePix/3
}

// Face returns the tag's font face for the tag's body
func (w *Tag) Face() font.Face {
	body, ok := w.Body.(facer)
	if !ok {
		return nil
	}
	return body.Face()
}

// SetFont sets the font face
func (w *Tag) SetFont(ft font.Face) {
	if w.Body == nil {
		return
	}
	body, ok := w.Body.(facer)
	if !ok {
		return
	}
	body.SetFont(ft)
	w.dirty = true
	w.Mark()
}
