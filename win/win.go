package win

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
	"sync"
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

	UserFunc func(*Win)
}

func (n *Win) Bounds() image.Rectangle {
	return image.Rectangle{n.Sp, n.Size()}
}

func (w *Win) Origin() int64 {
	return w.org
}
func (w *Win) Draw(dst draw.Image, r image.Rectangle, src image.Image, sp image.Point, op draw.Op){
	draw.Draw(dst,r,src,sp,op)
}
func (w *Win) Flush(r ...image.Rectangle) error{
	if len(r) == 0{
		w.Window().Upload(w.Sp.Add(w.b.Bounds().Min), w.b, w.b.Bounds())
		return nil
	}
	var wg sync.WaitGroup
	wg.Add(len(r))
	for _, r := range r{
		r := r
		go func(){
			w.Window().Upload(w.Sp.Add(r.Min), w.b, r)
			wg.Done()
		}()
	}
	wg.Wait()
	return nil
}
func (w *Win) StringBG(dst draw.Image, p image.Point, src image.Image, sp image.Point, ft *font.Font, s []byte, bg image.Image, bgp image.Point) int{
	return font.StringBG(dst,p,src,sp,ft,s,bg,bgp)
}

func New(dev *ui.Dev, sp, size, pad image.Point, ft *font.Font, cols frame.Color) *Win {
	r := image.Rectangle{pad, size}
	ed, _ := text.Open(text.NewBuffer())
	b := dev.NewBuffer(size)
	w := &Win{
		Node:     Node{Sp: sp, size: size, pad: pad},
		Dev:      dev,
		b:        b,
		Editor:   ed,
		UserFunc: func(w *Win) {},
	}
	w.Frame =  frame.NewDrawer(r, ft, b.RGBA(), cols, w)
	w.init()
	w.scrollinit(pad)

	return w
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
	return nil
}

func (w *Win) Resize(size image.Point) {
	b := w.NewBuffer(size)
	w.size = size
	w.b.Release()
	w.b = b
	r := image.Rectangle{w.pad, w.size} //.Inset(1)
	w.Frame = frame.NewDrawer(r, w.Frame.Font, w.b.RGBA(), w.Frame.Color, w, w.Frame.Flags())
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
	w.Frame = frame.NewDrawer(r, w.Frame.Font, w.b.RGBA(), w.Frame.Color, w, w.Frame.Flags())
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

func (w *Win) filldebug() {
	// Put
	fmt.Printf("lines/maxlines = %d/%d\n", w.Line(), w.MaxLine())
}

func (w *Win) Refresh() {
	w.Frame.Refresh()
	w.UserFunc(w)
	//w.Window().Upload(w.Sp, w.b, w.b.Bounds())
	w.Flush()
	w.dirty = false
}

// the old "Upload"
func (w *Win) Upload() {
//	w.Flush()
//	w.Window().Upload(w.Sp.Add(w.b.Bounds().Min), w.b, w.b.Bounds())
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
