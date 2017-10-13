package win

/*
 */

import (
	"fmt"
	"github.com/as/frame"
	"github.com/as/frame/font"
	"github.com/as/text"
	"github.com/as/ui"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/mouse"
	"image"
	"image/draw"
	"runtime"
	"sync"
)

const (
	MsgSize = 64 * 1024
)

func (w *Win) Dirty() bool {
	return w.dirty || (w.Frame != nil && w.Frame.Dirty())
}

type Node struct {
	Sp, size, pad image.Point
	dirty         bool
}

func (n Node) Size() image.Point {
	return n.Size()
}
func (n Node) Pad() image.Point {
	return n.Sp.Add(n.Size())
}

type Win struct {
	*frame.Frame
	text.Editor
	Node
	*ui.Dev
	b screen.Buffer
	ScrollBar
	org      int64
	Sq       int64
	inverted int

	donec    chan bool
	workc    chan image.Rectangle
	wg       sync.WaitGroup
	workerwg sync.WaitGroup

	UserFunc func(*Win)
}

func (n *Win) Bounds() image.Rectangle {
	return image.Rectangle{n.Sp, n.Size()}
}

func (w *Win) Origin() int64 {
	return w.org
}

func New(dev *ui.Dev, sp, size, pad image.Point, ft *font.Font, cols frame.Color) *Win {
	r := image.Rectangle{pad, size}
	ed, _ := text.Open(text.NewBuffer())
	b := dev.NewBuffer(size)
	w := &Win{
		Frame:    frame.New(r, ft, b.RGBA(), cols),
		Node:     Node{Sp: sp, size: size, pad: pad},
		Dev:      dev,
		b:        b,
		Editor:   ed,
		UserFunc: func(w *Win) {},
	}
	w.makeg()
	w.init()
	w.scrollinit(pad)
	return w
}

func (w *Win) workerspawn() {
	w.workerwg.Add(1)
	go func() {
		defer w.workerwg.Done()
		for {
			select {
			case <-w.donec:
				return
			case r := <-w.workc:
				w.Window().Upload(w.Sp.Add(r.Min), w.b, r)
				w.wg.Done()
			}
		}
	}()
}

func (w *Win) makeg() {
	ncpu := runtime.NumCPU()

	w.donec = make(chan bool)
	w.workc = make(chan image.Rectangle, ncpu*8)

	for i := 0; i < ncpu*2; i++ {
		w.workerspawn()
	}

	go func() {
		w.workerwg.Wait()
		close(w.workc)
	}()
}

func (w *Win) FuncInstall(fn func(*Win)) {
	if fn == nil {
		fn = func(w *Win) {}
	}
	w.UserFunc = fn
}

func (w *Win) Buffer() screen.Buffer {
	return w.b
}
func (w *Win) Size() image.Point {
	return w.size
}

func (w *Win) init() {
	w.Blank()
	w.Fill()
	q0, q1 := w.Dot()
	w.Select(q0, q1)
	w.Mark()
}

func (w Win) Loc() image.Rectangle {
	return image.Rectangle{w.Sp, w.Sp.Add(w.size)}
}

func (w *Win) Close() error {
	if w.Frame != nil {
		w.Frame.Close()
		w.Frame = nil
	}
	if w.b != nil {
		w.b.Release()
		w.b = nil
	}
	if w.Editor != nil {
		w.Editor.Close()
		w.Editor = nil
	}
	if w.donec != nil {
		close(w.donec)
		w.donec = nil
	}
	return nil
}

func (w *Win) Resize(size image.Point) {
	b := w.NewBuffer(size)
	w.size = size
	w.b.Release()
	w.b = b
	r := image.Rectangle{w.pad, w.size} //.Inset(1)
	w.Frame = frame.New(r, w.Frame.Font, w.b.RGBA(), w.Frame.Color)
	w.init()
	w.scrollinit(w.pad)
	w.Refresh()
}

func (w *Win) Move(sp image.Point) {
	w.Sp = sp
}

func (w *Win) SetFont(ft *font.Font) {
	if ft.Size() < 4 {
		return
	}
	r := image.Rectangle{w.pad, w.size}
	w.Frame = frame.New(r, ft, w.b.RGBA(), w.Frame.Color)
	w.Resize(w.size)
}

func (w *Win) NextEvent() (e interface{}) {
	switch e := w.Window().NextEvent().(type) {
	case mouse.Event:
		e.X -= float32(w.Sp.X)
		e.Y -= float32(w.Sp.Y)
		return e
	case interface{}:
		return e
	}
	return nil
}
func (w *Win) Send(e interface{}) {
	w.Window().Send(e)
}
func (w *Win) SendFirst(e interface{}) {
	w.Window().SendFirst(e)
}
func (w *Win) Blank() {
	if w.b == nil {
		return
	}
	r := w.b.RGBA().Bounds()
	draw.Draw(w.b.RGBA(), r, w.Color.Back, image.ZP, draw.Src)
	if w.Sp.Y > 0 {
		r.Min.Y--
	}
	w.Mark()
	w.drawsb()
	//w.upload()
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
		for j := 128; j-1 > 0 && p > 0; p-- {
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

func (w *Win) filldebug() {
	// Put
	fmt.Printf("lines/maxlines = %d/%d\n", w.Line(), w.MaxLine())
}

func (w *Win) Refresh() {
	w.Frame.Refresh()
	w.UserFunc(w)
	w.Window().Upload(w.Sp, w.b, w.b.Bounds())
	w.Flush()
	w.dirty = false
}

// Put
func (w *Win) Upload() {
	if !w.dirty {
		return
	}
	cache := append([]image.Rectangle{}, w.Cache()...)
	w.wg.Add(len(cache))
	for _, r := range w.Cache() {
		w.workc <- r
	}
	w.Flush()
	w.wg.Wait()
	w.dirty = false
}

// the old "Upload"
func (w *Win) zUpload() {
	if !w.dirty {
		return
	}
	var wg sync.WaitGroup
	wg.Add(len(w.Cache()))
	sp := w.Sp
	for _, r := range w.Cache() {
		go func(r image.Rectangle) {
			w.Window().Upload(sp.Add(r.Min), w.b, r)
			wg.Done()
		}(r)
	}
	wg.Wait()
	w.Flush()
	w.dirty = false
}

func (w *Win) ReadAt(off int64, p []byte) (n int, err error) {
	if off > w.Len() {
		return
	}
	return copy(p, w.Bytes()[off:w.Len()]), err
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
