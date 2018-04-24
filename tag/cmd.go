package tag

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"path/filepath"

	"github.com/as/path"
	"github.com/as/text"
	"github.com/as/text/action"
	"github.com/as/ui/win"
	"golang.org/x/mobile/event/mouse"
)

func Pt(e mouse.Event) image.Point {
	return image.Pt(int(e.X), int(e.Y))
}
func (t *Tag) Get(name string) {
	w := t.Body
	if w == nil {
		t.ctl <- fmt.Errorf("tag: window has no body for get request %q\n", name)
		return
	}
	if name == "" {
		t.fixtag("")
		return
	}

	abs := ""
	name, addr := action.SplitPath(name)
	if IsAbs(name) && path.Exists(name) {
		t.basedir = path.DirOf(name)
		abs = name
		t.fixtag(abs)
		t.getbody(abs, addr)
		return
	}
	abs = filepath.Join(t.basedir, name)
	t.fixtag(abs)
	t.getbody(abs, addr)
}

func (t *Tag) getbody(abs, addr string) {
	w := t.Body
	w.Delete(0, w.Len())
	w.Insert(t.readfile(abs), 0)
	w.Select(0, 0)
	w.SetOrigin(0, true)
	if addr != "" {
		t.ctl <- mustCompile(addr)
		//w.SendFirst(mustCompile(addr)) //TODO
	}
}

func (t *Tag) Put() (err error) {
	name := t.abs()
	if name == "" {
		return fmt.Errorf("no file")
	}
	t.ctl <- fmt.Errorf("Put %q", name)
	//	t.Window().Send(fmt.Errorf("Put %q", name)) // TODO
	t.Fs.Put(name, t.Body.Bytes())
	return nil
}
func Visible(w *win.Win, q0, q1 int64) bool {
	if q0 < w.Origin() {
		return false
	}
	if q1 > w.Origin()+w.Frame.Nchars {
		return false
	}
	return true
}

func Paste(w text.Editor, e mouse.Event) (int64, int64) {
	n, _ := Clip.Read(ClipBuf)
	s := fromUTF16(ClipBuf[:n])
	q0, q1 := w.Dot()
	if q0 != q1 {
		w.Delete(q0, q1)
		q1 = q0
	}
	w.Insert(s, q0)
	w.Select(q0, q0+int64(len(s)))
	return w.Dot()
}

func Rdsel(w text.Editor) string {
	q0, q1 := w.Dot()
	return string(w.Bytes()[q0:q1])
}

func Snarf(w text.Editor, e mouse.Event) {
	n := copy(ClipBuf, toUTF16([]byte(Rdsel(w))))
	io.Copy(Clip, bytes.NewReader(ClipBuf[:n]))
	q0, q1 := w.Dot()
	w.Delete(q0, q1)
	w.Select(q0, q0)
}
