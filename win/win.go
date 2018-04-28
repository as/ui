package win

import (
	"image"
	"image/draw"

	"github.com/as/font"
	"github.com/as/frame"
	"github.com/as/shiny/screen"
	"github.com/as/text"
	"github.com/as/ui"
	"github.com/as/ui/scroll"
)

type Facer func(int) font.Face

var (
	DefaultFaceHeight = 11
	DefaultMargin     = image.Pt(13, 3)
	MinRect           = image.Rect(0, 0, 10, 10)
)

var DefaultConfig = Config{
	Facer:  font.NewFace,
	Margin: DefaultMargin,
	Frame: &frame.Config{
		Face: font.NewFace(DefaultFaceHeight),
	},
}

type Config struct {
	Name string
	Facer
	Margin image.Point
	Frame  *frame.Config
	Editor text.Editor

	// Ctl is a channel provided by the window owner. It carries window messages
	// back to the creator. Valid types are event.Look and event.Cmd
	Ctl chan interface{}
}

type Win struct {
	*frame.Frame
	ui.Dev
	ctl chan interface{}

	b                screen.Buffer
	sp, size, margin image.Point

	org int64
	text.Editor
	dirty bool

	sb scroll.Bar
	Sq int64

	inverted int
	UserFunc func(*Win)
	Config   *Config
}

func (w *Win) Graphical() bool {
	return w.graphical()
}

func (w *Win) graphical() bool {
	return w != nil && w.Frame != nil && w.Dev != nil && w.b != nil && !w.Size().In(MinRect) && w.Size() != image.ZP
}

func (w *Win) Ctl() chan interface{} {
	return w.ctl
}

func New(dev ui.Dev, conf *Config) *Win {
	if conf == nil {
		c := DefaultConfig
		conf = &c
	}
	ed := conf.Editor
	if ed == nil {
		ed, _ = text.Open(text.NewBuffer())
	}
	w := &Win{
		Dev:      dev,
		ctl:      conf.Ctl,
		margin:   conf.Margin,
		Editor:   ed,
		UserFunc: func(w *Win) {},
		Config:   conf,
	}
	return w
}

func small(size image.Point) bool {
	return size.X == 0 || size.Y == 0 || size.In(MinRect)
}

// reallocimage releases the current image and replaces it with a fresh
// image of the given size. It returns false if the new size prevents the
// window from rendering graphics, meaning subsequent calls to
// w.Graphical returns false until a suitable image is allocated in its place
func (w *Win) reallocimage(size image.Point) bool {
	if w.b != nil {
		w.b.Release()
		w.b = nil
	}
	if small(size) {
		return false
	}
	b, err := w.NewBuffer(size)
	if err != nil {
		panic(size.String())
	}
	w.b = b
	return true
}

func (w *Win) Resize(size image.Point) {
	w.size = size
	if !w.reallocimage(w.size) {
		if w == nil {
			return
		}
		w.Frame = nil
		return
	}
	w.dirty = true
	r := image.Rectangle{w.margin, w.size}
	w.Frame = frame.New(w.b.RGBA(), r, w.Config.Frame)
	w.init()
	w.scrollinit(w.margin)
	w.Refresh()
}

func (w *Win) Dirty() bool {
	return w.dirty || (w.Frame != nil && w.Frame.Dirty())
}

func (w *Win) Buffer() screen.Buffer {
	return w.b
}
func (w *Win) Size() image.Point {
	return w.size
}

func (w *Win) Bounds() image.Rectangle {
	return image.Rectangle{w.sp, w.sp.Add(w.size)}
}

func (w Win) Loc() image.Rectangle {
	return w.Bounds()
}

func (w *Win) Origin() int64 {
	return w.org
}

func (w *Win) FuncInstall(fn func(*Win)) {
	if fn == nil {
		fn = func(w *Win) {}
	}
	w.UserFunc = fn
}

func (w *Win) init() {
	if w.graphical() {
		w.Blank()
		w.Fill()
	}
	q0, q1 := w.Dot()
	w.Select(q0, q1)
	w.Mark()
}

func (w *Win) Close() error {
	if w.Frame != nil {
		w.Frame.Close()
	}
	if w.b != nil {
		w.b.Release()
	}
	if w.Editor != nil {
		w.Editor.Close()
	}
	return nil
}

func (w *Win) Move(sp image.Point) {
	w.sp = sp
}

func (w *Win) SetFont(ft font.Face) {
	if ft.Height() < 4 {
		return
	}
	r := image.Rectangle{w.margin, w.size}
	w.Frame = frame.New(w.b.RGBA(), r, &frame.Config{Face: ft, Color: w.Frame.Color, Flag: w.Frame.Flags()})
	w.Resize(w.size)
}

func (w *Win) Visible() bool {
	return w.b != nil && w.Frame != nil && w.size != image.ZP
}

func (w *Win) Blank() {
	if !w.graphical() {
		return
	}
	r := w.minbounds()
	draw.Draw(
		w.b.RGBA(),
		r,
		w.Color.Back,
		image.ZP,
		draw.Src,
	)
	w.Mark()
	w.drawsb()
}
func (w *Win) Dot() (int64, int64) {
	return w.Editor.Dot()
}

// Swap swaps the primary foreground pallete with the highlighter pallete. It returns
// true if the color palletes are in their original order after the call to Swap.
func (w *Win) Swap() bool {
	w.Frame.Color.Palette, w.Frame.Color.Hi = w.Frame.Color.Hi, w.Frame.Color.Palette
	w.inverted++
	return w.inverted%2 == 0
}

func (w *Win) backNL(p int64, n int) int64 {
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
func (w *Win) Len() int64 {
	return w.Editor.Len()
}

func (w *Win) Refresh() {
	if !w.graphical() {
		return
	}
	w.Frame.Refresh()
	w.UserFunc(w)
	w.dirty = true
	w.Upload()
}

func (w Win) minbounds() image.Rectangle {
	return image.Rectangle{image.ZP, w.Bounds().Size()}.Union(w.b.Bounds())
}

func (w *Win) Upload() {
	w.dirty = true
	if !w.dirty || !w.graphical() {
		return
	}
	w.Window().Upload(w.sp, w.b, w.minbounds())
	w.Flush()
	w.dirty = false
}

func (w *Win) ReadAt(off int64, p []byte) (n int, err error) {
	if off > w.Len() {
		return
	}
	return copy(p, w.Bytes()[off:w.Len()]), err
}

func (w *Win) Readsel() []byte {
	q0, q1 := w.Dot()
	return w.Bytes()[q0:q1]
}

func (w *Win) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (w *Win) Bytes() []byte {
	return w.Editor.Bytes()
}

func (w *Win) Rdsel() []byte {
	q0, q1 := w.Dot()
	return w.Bytes()[q0:q1]
}
