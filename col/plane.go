package col

import (
	"image"
)

type Plane interface {
	Loc() image.Rectangle
	Move(image.Point)
	Resize(image.Point)
	Dirty() bool
	Refresh()
}
