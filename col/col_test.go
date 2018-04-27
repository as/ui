package col

import (
	"image"
	"os"
	"testing"

	etcher "github.com/as/etch"
	"github.com/as/ui"
	"github.com/as/ui/tag"
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

func testNewColHasNoEntries(t *testing.T, c *Col) {
	t.Helper()
	if len(c.List) != 0 {
		t.Fatalf("list has %v entries: %#v", len(c.List), c.List)
	}
}

func TestMoveNoSizeChange(t *testing.T) {
	r := image.Rect(55, 55, 155, 155)
	c := New(etch, r.Min, r.Size(), nil)
	testNonNil(t, c)
	testNewColHasNoEntries(t, c)

	sr := image.Rect(0, 0, 1000, 1000)
	etch.Blank()
	c.Upload(etch.Window())
	img0 := etch.Screenshot(sr)

	size0 := c.Loc().Size()
	c.Move(image.Pt(55, 55))
	testNewColHasNoEntries(t, c)

	size1 := c.Loc().Size()
	if size0 != size1 {
		t.Fatalf("size changed across calls to move: %s -move-> %s", size0, size1)
	}

	c.Upload(etch.Window())
	img1 := etch.Screenshot(sr)
	etcher.Assertf(t, img0, img1, "TestMoveNoSizeChange.png", "column size changed after move")
}

func TestAttachCoherence(t *testing.T) {
	r := image.Rect(55, 55, 155, 1024)
	tt := tag.New(etch, r.Min, r.Size(), nil)
	tt2 := tag.New(etch, r.Min, r.Size(), nil)
	c := New(etch, r.Min, r.Size(), nil)
	c.Attach(tt, 300)
	c.Attach(tt2, 500)

	y0 := c.Loc().Max.Y
	y1 := c.List[len(c.List)-1].Loc().Max.Y
	if y1 > y0 {
		t.Fatalf("extended y: %d < %d", y0, y1)
		etch.WritePNG("TestAttachCoherence.png")
	}
	c.Refresh()
	c.Tag.Insert([]byte("The"), 0)
	tt.Win.Insert([]byte("Quick"), 0)
	tt.Body.Insert([]byte("Brown"), 0)
	tt2.Body.Insert([]byte("Fox"), 0)
	c.Move(image.Pt(500, 1))
	c.Refresh()
	c.Upload(etch.Window())
	c.Move(image.Pt(700, 10))
	c.Resize(c.Loc().Size().Add(image.Pt(100, 0)))
	if c.Loc().Size().Y != r.Size().Y {
		t.Fatalf("attach extended y-axis: %d -> %d", r.Size().Y, c.Loc().Size().Y)
	}
	c.Refresh()
	c.Upload(etch.Window())
}

func TestAttach(t *testing.T) {
	r := image.Rect(55, 55, 155, 1555)
	tt := tag.New(etch, r.Min, r.Size(), nil)
	tt2 := tag.New(etch, r.Min, r.Size(), nil)
	c := New(etch, r.Min, r.Size(), nil)
	c.Attach(tt, 1555)
	c.Attach(tt2, 700)
	c.Refresh()
	c.Tag.Insert([]byte("The"), 0)
	tt.Win.Insert([]byte("Quick"), 0)
	tt.Body.Insert([]byte("Brown"), 0)
	tt2.Body.Insert([]byte("Fox"), 0)
	c.Move(image.Pt(500, 1))
	c.Refresh()
	c.Upload(etch.Window())
	c.Move(image.Pt(700, 10))
	c.Resize(c.Loc().Size().Add(image.Pt(100, 0)))
	c.Refresh()
	c.Upload(etch.Window())
	etch.WritePNG("TestAttach.png")
}

func TestNew(t *testing.T) {
	t.Skip()
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
		c.Upload(etch.Window())
	}
	img1 := etch.Screenshot(r)
	//	t.Skip("the scrollbar colors dont match; thats fine for now")
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