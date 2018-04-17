// Package ui is a wrapper around a graphical driver like shiny
// Basically, I don't like having two extra arguments per function
// so this is just an initialization package as well as a struct
// to hold the screen and window pointers.
package ui

import (
	"image"

	"github.com/as/shiny/driver"
	"github.com/as/shiny/screen"
)

type Item interface {
	Buffer() screen.Buffer
	Send(e interface{})
	SendFirst(e interface{})
	NextEvent() (e interface{})
}

type Win interface {
	Item
	Blank()
	Bounds() image.Rectangle
	Bytes() []byte
	Dirty() bool
	Fill()
	Len() int64
	Refresh()
	Size() image.Point
	Upload()
	Resize(size image.Point)
	Move(sp image.Point)
}

type Dev struct {
	scr    screen.Screen
	events screen.Window
	killc  chan bool
}

func Init(opts *screen.NewWindowOptions) (dev *Dev, err error) {
	errc := make(chan error)
	go func(errc chan error) {
		driver.Main(func(scr screen.Screen) {
			wind, err := scr.NewWindow(opts)
			if err != nil {
				errc <- err
			}
			dev = &Dev{scr, wind, make(chan bool)}
			errc <- err
			<-dev.killc
		})
	}(errc)
	return dev, <-errc
}
func (d *Dev) Screen() screen.Screen { return d.scr }
func (d *Dev) Window() screen.Window { return d.events }
func (d *Dev) NewBuffer(size image.Point) screen.Buffer {
	b, err := d.scr.NewBuffer(size)
	if err != nil {
		panic(err)
	}
	return b
}

type Node struct {
	sp, size image.Point
}

func (n Node) Move(pt image.Point) {
	n.sp = pt
}

func (n Node) Resize(pt image.Point) {
	n.size = pt
}

func (n Node) Size() image.Point {
	return n.size
}

func (n Node) Pad() image.Point {
	return n.sp.Add(n.Size())
}
