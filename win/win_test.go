package win

import (
	"image"
	"os"
	"testing"

	"github.com/as/ui"
)

var etch ui.Dev

func TestMain(m *testing.M) {
	etch = ui.NewEtch()
	os.Exit(m.Run())
}

func wantshape(t *testing.T, w *Win, want image.Rectangle) {
	t.Helper()
	if w.Bounds() != want {
		t.Fatalf("wantshape: %s, have %s", want, w.Bounds())
	}
	if w.Bounds() != w.Loc() {
		t.Fatalf("wantshape: loc %s != bounds %s", w.Loc(), w.Bounds())
	}
}
func wantbytes(t *testing.T, w *Win, text interface{}) {
	want := ""
	v, _ := text.([]byte)
	if v == nil {
		want = text.(string)
	} else {
		want = string(v)
	}

	t.Helper()
	if have := string(w.Readsel()); have != want {
		t.Fatalf("have string: %s, want %s", have, want)
	}
}
func wantdot(t *testing.T, w *Win, q0, q1 int64) {
	t.Helper()
	r0, r1 := w.Dot()
	if r0 != q0 || r1 != q1 {
		t.Fatalf("bad dot: have [%d:%d] want [%d:%d]", r0, r1, q0, q0)
	}
}

func TestNewt(t *testing.T) {
	w := New(etch, image.ZP, image.Pt(1000, 1000), nil)
	if w == nil {
	}
	if w.Loc() != image.Rect(0, 0, 1000, 1000) {
		t.Fatalf("pure zero tag has non zero location: %s", w.Loc())
	}
	w.Insert([]byte("hhello"), 0)
	w.Resize(image.Pt(0, 500))
	w.Delete(0, 1)
	w.Select(0, 1)
	w.Dot()
	w.SetOrigin(4, true)
	w.SetOrigin(4, false)
	if string(w.Bytes()) != "hello" {
		t.Fatalf("filename: want %q, have %q", "hello", string(w.Bytes()))
	}
	w.Resize(image.Pt(1000, 1000))
	w.Move(image.Pt(55, 55))
	wantshape(t, w, image.Rect(55, 55, 1055, 1055))
	w.Resize(image.Pt(500, 500))
}

func TestLocation(t *testing.T) {
	r := image.ZR
	var x0, y0, x1, y1 int
	for y0 = 0; y0 < 100; y0 += 5 {
		for x0 = 0; x0 < 101; x0 += 7 {
			for y1 = 0; y1 < 103; y1 += 11 {
				for x1 = 0; x1 < 107; x1 += 13 {
					r = image.Rect(x0, y0, x1, y1)
					w := New(etch, r.Min, r.Size(), nil)
					if w == nil {
						t.Fatalf("%v: window is nil", r)
					}
					if w.Loc() != r {
						t.Fatalf("%v: bad location: %v", r, w.Loc())
					}

					r2 := r.Add(image.Pt(13, 13))
					w.Move(r2.Min)
					wantshape(t, w, r2)

					w.Resize(image.Pt(0, 0))
					wantshape(t, w, image.Rectangle{r2.Min, r2.Min})

				}
			}
		}
	}
}

func testStatelessCrashers(t *testing.T, w *Win) {
	t.Helper()
	w.Dot()
	w.Blank()
	w.Bounds()
	w.Loc()
	w.Upload()
}

func TestHiddenUnhide(t *testing.T) {
	var in = []byte("hello")
	w := New(etch, image.ZP, image.ZP, nil)

	wantshape(t, w, image.ZR)
	w.Insert(in, 0)
	w.Select(0, 5)
	q0, q1 := w.Dot()
	wantdot(t, w, q0, q1)
	wantbytes(t, w, in)

	testStatelessCrashers(t, w)

	w.Resize(image.Pt(500, 500))
	w.Frame.Insert(in, 0)

	wantshape(t, w, image.Rect(0, 0, 500, 500))
	w.Select(0, 5)
	q0, q1 = w.Dot()
	wantdot(t, w, q0, q1)
	wantbytes(t, w, in)

	w.pngwrite("TestHiddenUnhide.have.png")
}
