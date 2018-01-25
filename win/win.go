package win

import (
	"image"
	"image/draw"

	"github.com/as/frame"
	"github.com/as/shrew"
	"github.com/as/text"
	"golang.org/x/image/font"
)

type Win struct {
	*frame.Frame
	text.Editor
	ScrollBar
	org      int64
	Sq       int64
	inverted int

	buffer bool
	c      shrew.Client
	cacher *cacher
}

type Config struct {
	Name   string
	Flag   int
	Pad    image.Point
	Face   font.Face
	Color  *frame.Color
	Drawer frame.Drawer
	Editor text.Editor
}

func (c *Config) check() {
	if c.Face == nil {
		c.Face = frame.NewGoMono(11)
	}
	if c.Pad == image.ZP {
		c.Pad = image.Pt(15, 15)
	}
	if c.Editor == nil {
		c.Editor, _ = text.Open(text.NewBuffer())
	}
}

func New(c shrew.Client, conf *Config) *Win {
	if conf == nil {
		conf = &Config{}
	}
	conf.check()
	r := c.Bounds()
	r.Min.X += conf.Pad.X
	r.Min.Y += conf.Pad.Y
	cacher := &cacher{Bitmap: c}
	w := &Win{
		c:      c,
		Editor: conf.Editor,
		cacher: cacher,
		Frame: frame.New(c, r, &frame.Config{
			Color:  conf.Color,
			Font:   conf.Face,
			Flag:   conf.Flag,
			Drawer: cacher,
		}),
	}
	w.init()
	w.scrollinit(r)
	return w
}

func (w *Win) init() {
	w.Blank()
	w.Fill()
	q0, q1 := w.Dot()
	w.Select(q0, q1)
}
func (w *Win) Blank() {
	r := w.c.Bounds()
	w.c.Draw(w.c, r, w.Color.Back, image.ZP, draw.Src)
	w.drawsb()
}
func (w *Win) Dot() (int64, int64) {
	return w.Editor.Dot()
}
func (w *Win) BackNL(p int64, n int) int64 {
	if n == 0 && p > 0 && w.Bytes()[p-1] != '\n' {
		n = 1
	}
	for i := n; i > 0 && p > 0; {
		i--
		p--
		if p == 0 {
			break
		}
		for j := 512; j-1 > 0 && p > 0; p-- {
			j--
			if p-1 < 0 || p-1 > w.Len() || w.Bytes()[p-1] == '\n' {
				break
			}
		}
	}
	return p
}
func (w *Win) Resize(sp image.Point) {

}
func (w *Win) Move(sp image.Point) {

}
func (w *Win) Close() error {
	return nil
}
func (w *Win) Len() int64 {
	return w.Editor.Len()
}
func (w *Win) Refresh() {
	w.Frame.Refresh()
}
func (w *Win) Bytes() []byte {
	return w.Editor.Bytes()
}
