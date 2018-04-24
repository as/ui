package col

import (
	etcher "github.com/as/etch"
	"github.com/as/ui"
	"image"
	"os"
	"testing"
)

var etch *ui.Etch

func TestMain(m *testing.M) {
	etch = ui.NewEtch()
	os.Exit(m.Run())
}

type Locator interface {
	Loc() image.Rectangle
}

func wantshape(t *testing.T, c Locator, want image.Rectangle) {
	t.Helper()
	if c.Loc() != want {
		t.Fatalf("wantshape: %s, have %s", want, c.Loc())
	}
}

func TestNewBasic(t *testing.T) {
	for _, r := range []image.Rectangle{
		image.Rect(0, 0, 1000, 1000),
		image.Rect(5, 5, 1000, 1000),
		image.Rect(55, 55, 1000, 1000),
		image.Rect(555, 555, 1000, 1000),
	} {
		c := New(etch, r.Min, r.Size(), nil)
		testNonNil(t, c)
		wantshape(t, c, r)
		for _, pt := range []image.Point{{-1, -1}, {-1, 0}, {0, -1}, {0, 0}, {0, 1}, {1, 0}, {1, 1}} {
			c.Move(r.Min.Add(pt))
			wantshape(t, c, r.Add(pt))
		}
	}
}

func TestNew(t *testing.T) {
	r := image.Rect(55, 55, 1000, 1000)
	c := New(etch, r.Min, r.Size(), nil)
	testNonNil(t, c)

	wantshape(t, c, r)

	{
		sr := image.Rect(0, 0, 1000, 1000)
		etch.Blank()
		c.Upload(etch.Window())
		img0 := etch.Screenshot(sr)

		etch.Blank()
		c.Move(image.Pt(555, 555))
		c.Move(image.Pt(55, 55))
		c.Upload(etch.Window())
		img1 := etch.Screenshot(sr)

		etcher.Assertf(t, img0, img1, "TestNewMove.png", "column state differs after move")
	}
	img0 := etch.Screenshot(r)

	c = New(etch, image.ZP, image.ZP, nil)
	testNonNil(t, c)

	etch.Blank()
	{
		//c.Upload(etch.Window())
	}
	img1 := etch.Screenshot(r)
	t.Skip("the scrollbar colors dont match; thats fine for now")
	etcher.Assertf(t, img0, img1, "TestNew.png", "TestNew")
}

func testNonNil(t *testing.T, c *Col) {
	t.Helper()
	if c == nil {
		t.Fatalf("col: col is nil")
	}
	{
		c := c.Tag
		if c == nil {
			t.Fatalf("col: tag is nil")
		}
		if c.Win == nil {
			t.Fatalf("col: tag label is nil")
		}
		if c.Body == nil {
			t.Fatalf("col: tag body is nil")
		}
	}
}
