package col

import (
	"image"

	"github.com/as/ui/win"
)

type Plane interface {
	Loc() image.Rectangle
	Move(image.Point)
	Resize(image.Point)
	Dirty() bool
	Refresh()
}

func (col *Col) Dirty() bool {
	for _, v := range col.List {
		if v.Dirty() {
			return true
		}
	}
	return false
}

func (c *Col) Label() *win.Win { return c.Tag.Win }
func (c *Col) Kid(n int) Plane {
	return c.List[n]
}
func (col *Col) Kids() []Plane {
	return col.List
}

func (c *Col) Len() int {
	return len(c.List)
}

func (col *Col) Refresh() {
	col.Tag.Refresh()
	for _, v := range col.List {
		v.Refresh()
	}
}

type Named interface {
	Plane
	FileName() string
}

func (col *Col) Lookup(pid interface{}) Plane {
	kids := col.Kids()
	if len(kids) == 0 {
		return nil
	}
	switch pid := pid.(type) {
	case int:
		if pid >= len(kids) {
			pid = len(kids) - 1
		}
		return col.Kids()[pid]
	case string:
		for i, v := range col.Kids() {
			if v, ok := v.(Named); ok {
				if v.FileName() == pid {
					return col.Kids()[i]
				}
			}
		}
	case image.Point:
		return ptInAny(pid, col.Kids()...)
	case interface{}:
		panic("")
	}
	return nil
}

func ptInAny(pt image.Point, list ...Plane) (x Plane) {
	for i, w := range list {
		if ptInPlane(pt, w) {
			return list[i]
		}
	}
	return nil
}

func ptInPlane(pt image.Point, p Plane) bool {
	if p == nil {
		return false
	}
	return pt.In(p.Loc())
}
