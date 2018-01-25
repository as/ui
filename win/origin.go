package win

import (
	"github.com/as/text"
)

func (w *Win) Origin() int64 {
	return w.org
}

func (w *Win) SetOrigin(org int64, exact bool) {
	org = clamp(org, 0, w.Len())
	if org == w.org {
		return
	}
	//	w.Mark()
	if org > 0 && !exact {
		for i := 0; i < 2048 && org < w.Len(); i++ {
			if w.Bytes()[org] == '\n' {
				org++
				break
			}
			org++
		}
	}
	w.setOrigin(clamp(org, 0, w.Len()))
}

func (w *Win) setOrigin(org int64) {
	if org == w.org {
		return
	}
	fl := w.Frame.Len()
	switch text.Region5(org, org+fl, w.org, w.org+fl) {
	case -1:
		// Going down a bit
		w.Frame.Insert(w.Bytes()[org:org+(w.org-org)], 0)
		w.org = org
	case -2, 2:
		w.Frame.Delete(0, w.Frame.Len())
		w.org = org
		w.Fill()
	case 1:
		// Going up a bit
		w.Frame.Delete(0, org-w.org)
		w.org = org
		w.Fill()
		//w.fixEnd()

	case 0:
		panic("never happens")
	}
	q0, q1 := w.Dot()
	w.drawsb()
	w.Select(q0, q1)
}

func (w *Win) fixEnd() {
	fr := w.Frame.Bounds()
	if pt := w.PointOf(w.Frame.Len()); pt.Y != fr.Max.Y {
		w.Paint(pt, fr.Max, w.Frame.Color.Palette.Back)
	}
}
