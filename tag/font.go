package tag

import (
	"github.com/as/font"
	"github.com/as/ui/win"
)

// Height returns the recommended minimum pixel height for a tag label
// given the face height in pixels.
func Height(facePix int) int {
	if facePix == 0 {
		facePix = DefaultConfig.FaceHeight
	}
	return facePix + facePix/2 + facePix/3
}

// SetFont sets the font face
func (w *Tag) SetFont(ft font.Face) {
	body := w.Body.(*win.Win)
	if body == nil {
		return
	}
	if ft.Height() < 3 || w.Body == nil {
		return
	}
	body.SetFont(ft)
	w.dirty = true
	w.Mark()
	body.Refresh()
}
