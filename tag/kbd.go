package tag

import (
	"github.com/as/font"
	"github.com/as/text/find"
	"github.com/as/text/kbd"
	"github.com/as/ui/win"
	"golang.org/x/mobile/event/key"
)

// Kbd processes the given key press event. act is the currently active window
// in the tag, but it should really be removed because it's hideous.
//
// TODO(as): Find some other way to determine which piece of the window gets
// the key.
func (t *Tag) Kbd(e key.Event, act Window) {
	if e.Direction == 2 {
		return
	}
	if e.Code == key.CodeI && e.Modifiers == key.ModControl {
		runGoImports(t, e)
		return
	}
	switch e.Code {
	case key.CodeEqualSign, key.CodeHyphenMinus:
		if e.Modifiers == key.ModControl {
			win, _ := t.Body.(*win.Win)
			if win == nil {
				return
			}
			size := win.Frame.Face.Height()
			if key.CodeHyphenMinus == e.Code {
				size -= 1
			} else {
				size += 1
			}
			if size < 3 {
				size = 6
			}
			t.SetFont(font.NewFace(size))
			//t.SetFont(t.Body.Frame.Face.NewSize(size))
			return
		}
	}
	ntab := int64(-1)
	if (e.Rune == '\n' || e.Rune == '\r') && act == t.Body {
		q0, q1 := act.Dot()
		if q0 == q1 {
			p := act.Bytes()
			l0, _ := find.Findlinerev(p, q0, 0)
			ntab = find.Accept(p, l0, []byte{'\t'})
			ntab -= l0 + 1
		}
	}
	kbd.SendClient(act, e)
	for ntab >= 0 {
		e.Rune = '\t'
		kbd.SendClient(act, e)
		ntab--
	}
	t.Mark()
}
