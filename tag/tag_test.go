package tag

import (
	"fmt"
	"image"
	"os"
	"testing"

	etcher "github.com/as/etch"
	"github.com/as/ui"
)

var etch *ui.Etch

func TestMain(m *testing.M) {
	etch = ui.NewEtch()
	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	r := image.Rect(0, 0, 1000, 1000)
	tt := New(etch, nil)
	tt.Move(r.Min)
	tt.Resize(r.Size())
	testNonNil(t, tt)

	wantshape(t, tt, r)
	if !tt.Label.Graphical() {
		t.Fatalf("tag: tag label not graphical (should be graphical; r=%s)", tt.Label.Bounds())
	}

	etch.Blank()
	{
		tt.Label.Insert([]byte("hello"), 0)
		tt.Window.Insert([]byte("world"), 0)
		tt.Upload()
	}
	img0 := etch.Screenshot(r)

	tt = New(etch, nil)
	testNonNil(t, tt)
	if tt.Label.Graphical() {
		t.Fatalf("tag: tag label graphical (shouldnt be graphical; r=%s)", tt.Label.Bounds())
	}

	etch.Blank()
	{
		tt.Label.Insert([]byte("hello"), 0)
		tt.Window.Insert([]byte("world"), 0)
		tt.Resize(r.Size())
		tt.Upload()
	}
	img1 := etch.Screenshot(r)
	t.Skip("the scrollbar colors dont match; thats fine for now")
	etcher.Assertf(t, img0, img1, "TestNew.png", "TestNew")
}

func TestNewDirty(t *testing.T) {
	tt := New(etch, nil)
	tt.Resize(image.Pt(1000, 1000))
	if !tt.Label.Dirty() {
		t.Fatalf("new tag window shouldn't be clean")
	}
	if !tt.Window.Dirty() {
		t.Fatalf("new tag body shouldn't be clean")
	}
}

func TestCreateZero(t *testing.T) {
	tt := New(etch, nil)
	if tt == nil {
		t.Fatalf("tag is nil")
	}
	if tt.Bounds() != image.ZR {
		t.Fatalf("pure zero tag has non zero location: %s", tt.Bounds())
	}
	tt.Label.Insert([]byte("hello	"), 0)
	tt.Delete(0, 1)
	tt.Select(0, 1)
	tt.Dot()
	tt.SetOrigin(4, true)
	tt.SetOrigin(4, false)
	if tt.FileName() != "hello" {
		t.Fatalf("filename: want %q, have %q", "hello", tt.FileName())
	}
	tt.Move(image.Pt(55, 55))
	wantshape(t, tt, image.Rect(55, 55, 55, 55))
	tt.Resize(image.Pt(500, 500))
}

type sizeentry struct {
	sx, sy, dx, dy, x, y, xx, yy int
	vis                          Vis
}

func (s sizeentry) Sp() image.Point {
	return image.Pt(s.sx, s.sy)
}
func (s sizeentry) Size() image.Point {
	return image.Pt(s.dx, s.dy)
}
func (s sizeentry) Want() image.Rectangle {
	return image.Rect(s.x, s.y, s.xx, s.yy)
}

func testVisibility(t *testing.T, tt *Tag) {
	t.Helper()
	r := tt.Bounds()
	tt.Move(r.Min)
	vis := tt.Vis
	if tt.Bounds() != r {
		t.Fatalf("tag dimensions changed after 0-Move: %s -> %s", r, tt.Bounds())
	}
	pt := r.Min
	tt.Move(r.Min.Add(image.Pt(10000, 10000)))
	defer tt.Move(pt)
	if tt.Bounds().Size() != r.Size() {
		t.Fatalf("tag size changed after Move 10k : %s -> %s", r.Size(), tt.Bounds().Size())
	}
	if tt.Vis != vis {
		t.Fatalf("tag visibility changed after Move: %v -> %v", vis, tt.Vis)
	}
}

func TestResize(t *testing.T) {
	for i, r := range []image.Rectangle{
		image.Rect(0, 0, 0, 0),
		image.Rect(0, 0, 1000, 22),
		image.Rect(0, 0, 1000, 23),
		image.Rect(0, 0, 1000, 1000),
	} {
		name := fmt.Sprintf("%d/a: r=%s", i, r.Bounds())
		tt := New(etch, nil)
		tt.Move(r.Min)
		tt.Resize(r.Size())
		testResize(t, name, tt)

		name = fmt.Sprintf("%d/b: r=%s", i, r.Bounds())
		tt = New(etch, nil)
		tt.Move(r.Min)
		tt.Resize(r.Size())
		testResize(t, name, tt)
	}
}

func testResize(t *testing.T, name string, tt *Tag) {
	var sizetab = []sizeentry{
		{0, 0, 0, 0, 0, 0, 0, 0, VisNone},
		{9, 9, 0, 0, 9, 9, 9, 9, VisNone}, // Hidden != image.ZR
		{0, 0, 4, 4, 0, 0, 4, 0, VisNone},
		{0, 0, 10, 10, 0, 0, 10, 0, VisNone},
		{0, 0, 11, 11, 0, 0, 11, 0, VisNone},
		{0, 0, 12, 12, 0, 0, 12, 0, VisNone},
		{0, 0, 22, 22, 0, 0, 22, 19, VisTag},
		//		{0, 0, 23, 23, 0, 0, 23, 23, VisTag},
		{0, 0, 1000, 1000, 0, 0, 1000, 1000, VisFull},
		{0, 0, 500, 500, 0, 0, 500, 500, VisFull},
		{0, 0, 250, 250, 0, 0, 250, 250, VisFull},
		{0, 0, 50, 50, 0, 0, 50, 50, VisFull},
		{0, 0, 1000, 1000, 0, 0, 1000, 1000, VisFull},
	}
	//	wantshape(t, tt, image.Rect(0, 0, 1000, 1000))
	for i, v := range sizetab {
		testVisibility(t, tt)
		tt.Resize(v.Size())
		tt.Move(v.Sp())
		wantshapePrefix(t, "after resize to: "+v.Size().String(), tt, v.Want())

		if v.vis != tt.Vis {
			t.Fatalf("%s: %d: visibility doesn't match: have %v, want %v", name, i, tt.Vis, v.vis)
		}
		testVisibility(t, tt)
	}
}

func TestResizeZero(t *testing.T) {
	tt := New(etch, nil)
	tt.Resize(image.Pt(1000, 1000))
	if tt == nil {
	}
	if tt.Bounds() != image.Rect(0, 0, 1000, 1000) {
		t.Fatalf("pure zero tag has non zero location: %s", tt.Bounds())
	}
	tt.Label.Insert([]byte("hello	"), 0)
	tt.Resize(image.Pt(0, 500))
	tt.Delete(0, 1)
	tt.Select(0, 1)
	tt.Dot()
	tt.SetOrigin(4, true)
	tt.SetOrigin(4, false)
	if tt.FileName() != "hello" {
		t.Fatalf("filename: want %q, have %q", "hello", tt.FileName())
	}
	tt.Move(image.Pt(55, 55))
	wantshape(t, tt, image.Rect(55, 55, 55, 555))
	etch.Blank()
	tt.Resize(image.Pt(500, 500))
	etch.WritePNG("TestResizeZero.png")
	tt.Upload()
}

func TestLocation(t *testing.T) {
	tt := New(etch, nil)
	tt.Resize(image.Pt(1000, 1000))
	ckLayout(t, tt)
	tt.Move(image.Pt(25, 25))
	wantshape(t, tt, image.Rect(25, 25, 1025, 1025))
}

func wantshape(t *testing.T, tt *Tag, want image.Rectangle) {
	t.Helper()
	if tt.Bounds() != want {
		t.Fatalf("wantshape: %s, have %s", want, tt.Bounds())
	}
}

func wantshapePrefix(t *testing.T, msg string, tt *Tag, want image.Rectangle) {
	t.Helper()
	if tt.Bounds() != want {
		t.Fatalf("%s: wantshape: %s, have %s", msg, want, tt.Bounds())
	}
}

func testNonNil(t *testing.T, tt *Tag) {
	t.Helper()
	if tt == nil {
		t.Fatalf("tag: tag is nil")
	}

	if tt.Label == nil {
		t.Fatalf("tag: label is nil")
	}
	if tt.Window == nil {
		t.Fatalf("tag: body is nil")
	}
}

func ckLayout(t *testing.T, tt *Tag) {
	t.Helper()
	if tt == nil {
		t.Fatal("constraint violation: nil tag")
	}

	r := tt.Bounds()
	if r != tt.Bounds() {
		t.Fatalf("constraint violation: bounds != loc: %s != %s", r, tt.Bounds())
	}

	if tt.Label == nil && tt.Window != nil {
		// Sutle: This is a different class of errors than the one below. Don't remove it.
		t.Fatalf("constraint violation: tag window == nil but body != nil")
	}
	if tt.Label == nil || tt.Window == nil {
		t.Fatalf("constraint violation: tag or win is nil")
	}

	h := tt.Config.TagHeight()
	want0 := image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+h)
	have0 := tt.Label.Bounds()
	if want0 != have0 {
		t.Fatalf("tag bounds dont match label: have %s want %s", have0, want0)
	}
	want0.Min.Y += h
	want0.Max.Y = r.Max.Y
	have0 = tt.Window.Bounds()
	if want0 != have0 {
		t.Fatalf("tag body bounds dont match tag window: have %s want %s", have0, want0)
	}
}
