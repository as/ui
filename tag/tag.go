package tag

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"log"
	"os"
	"strings"
	//	"time"
	"path/filepath"
	"sync"

	"github.com/as/srv/fs"

	"github.com/as/edit"
	"github.com/as/event"
	"github.com/as/font"
	"github.com/as/frame"
	"github.com/as/path"
	"github.com/as/text"
	"github.com/as/text/action"
	"github.com/as/text/find"
	"github.com/as/text/kbd"
	mus "github.com/as/text/mouse"
	"github.com/as/ui"
	"github.com/as/ui/win"
	//"github.com/as/worm"
	"github.com/as/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/mouse"
)

func p(e mouse.Event) image.Point {
	return image.Pt(int(e.X), int(e.Y))

}

var (
	Buttonsdown = 0
)

type Tag struct {
	*win.Win
	Body      *win.Win
	Scrolling bool
	fs.Fs

	sp image.Point

	dirty  bool
	r0, r1 int64
	escR   image.Rectangle

	ctl chan<- interface{}

	basedir string
}

func (w *Tag) SetFont(ft font.Face) {
	if ft.Height() < 3 || w.Body == nil {
		return
	}
	w.Body.SetFont(ft)
	w.dirty = true
	w.Mark()
	w.Body.Refresh()
}

func (t *Tag) Dirty() bool {
	return t.dirty || t.Win.Dirty() || (t.Body != nil && t.Body.Dirty())
}

func (t *Tag) Mark() {
	t.dirty = true
	t.Win.Mark()
}

func (t *Tag) Loc() image.Rectangle {
	r := t.Win.Loc()
	if t.Body != nil {
		r.Max.Y += t.Body.Loc().Dy()
	}
	return r
}

// TagSize returns the size of a tag given the font
func TagSize(ft font.Face) int {
	return ft.Dy() + ft.Dy()/2
}

// TagPad returns the padding for the tag given the window's padding
// always returns an x-aligned point
func TagPad(wpad image.Point) image.Point {
	return image.Pt(wpad.X, 3)
}

type Config struct {
	Facer      func(int) font.Face
	FaceHeight int
	Margin     image.Point
	Color      [3]frame.Color
	Tag        *win.Config
	Body       *win.Config
	Ctl        chan interface{}
	Filesystem fs.Fs
}

func (c *Config) TagConfig() *win.Config {
	return &win.Config{
		Ctl:    c.Ctl,
		Facer:  c.Facer,
		Margin: image.Pt(c.Margin.X, 3),
		Frame: &frame.Config{
			Color: c.Color[0],
			Face:  c.Facer(c.FaceHeight),
		},
	}
}
func (c *Config) WinConfig() *win.Config {
	return &win.Config{
		Ctl:    c.Ctl,
		Facer:  c.Facer,
		Margin: c.Margin,
		Frame: &frame.Config{
			Color: c.Color[1],
			Face:  c.Facer(c.FaceHeight),
		},
	}
}

func New(dev *ui.Dev, sp, size image.Point, conf *Config) *Tag {
	if conf == nil {
		conf = &Config{
			FaceHeight: 11,
			Facer:      font.NewFace,
			Margin:     image.Pt(15, 15),
			Color: [3]frame.Color{
				0: frame.ATag1,
				1: frame.A,
			},
		}
	}
	if conf.Ctl == nil {
		panic("ctl cant be nil")
	}
	if conf.Filesystem == nil {
		conf.Filesystem = &fs.Local{}
	}
	tconf := conf.TagConfig()
	tagY := TagSize(tconf.Frame.Face.(font.Face))
	wtag := win.New(dev, sp, image.Pt(size.X, tagY), tconf)

	sp = sp.Add(image.Pt(0, tagY))
	size = size.Sub(image.Pt(0, tagY))
	if size.Y < tagY {
		return &Tag{sp: sp, Win: wtag, Body: nil, ctl: conf.Ctl}
	}

	w := win.New(dev, sp, size, conf.WinConfig())

	wd, _ := os.Getwd()
	return &Tag{sp: sp, Win: wtag, Body: w, basedir: wd, ctl: conf.Ctl, Fs: conf.Filesystem}
}

func (t *Tag) Move(pt image.Point) {
	t.Win.Move(pt)
	if t.Body == nil {
		return
	}
	pt.Y += t.Win.Loc().Dy()
	t.Body.Move(pt)
}

func (t *Tag) Resize(pt image.Point) {
	var wg sync.WaitGroup
	defer wg.Wait()

	dy := TagSize(t.Win.Face)
	if pt.X < dy || pt.Y < dy {
		println("bad size request:", pt.String())
		return
	}
	wg.Add(1)
	go func() {
		t.Win.Resize(image.Pt(pt.X, dy))
		wg.Done()
	}()

	if t.Body != nil {
		pt := pt
		pt.Y -= dy
		wg.Add(1)
		go func() {
			t.Body.Resize(pt)
			wg.Done()
		}()
	}
}

func mustCompile(prog string) *edit.Command {
	p, err := edit.Compile(prog)
	if err != nil {
		log.Printf("tag.go:/mustCompile/: failed to compile %q\n", prog)
		return nil
	}
	return p
}

func (t *Tag) Open(basepath, title string) {
	t.basedir = path.DirOf(basepath)
	println(title)
	t.Get(title)
}

func (t *Tag) Close() (err error) {
	if t.Body != nil {
		err = t.Body.Close()
	}
	if t.Win != nil {
		err = t.Win.Close()
	}
	return err
}

func (t *Tag) Dir() string {
	x := path.DirOf(t.FileName())
	if IsAbs(x) {
		return x
	}
	return filepath.Join(t.basedir, x)
}

func (t *Tag) fixtag(abs string) {
	wtag := t.Win
	p := wtag.Bytes()
	maint := find.Find(p, 0, []byte{'|'})
	if maint == -1 {
		maint = int64(len(p))
	}
	wtag.Delete(0, maint+1)
	wtag.InsertString(abs+"\tPut Del |", 0)
	wtag.Refresh()
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
	if !path.Exists(abs) {
		//
	}
	t.fixtag(abs)
	t.getbody(abs, addr)
}

type GetEvent struct {
	Basedir string
	Name    string
	Addr    string
	IsDir   bool
}

func (t *Tag) abs() string {
	name := t.FileName()
	if !IsAbs(name) {
		name = filepath.Join(t.basedir, name)
	}
	return name
}

func (t *Tag) FileName() string {
	if t == nil || t.Win == nil {
		return ""
	}
	name, err := bufio.NewReader(bytes.NewReader(t.Win.Bytes())).ReadString('\t')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(name)
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
func pt(e mouse.Event) image.Point {
	return image.Pt(int(e.X), int(e.Y))
}
func (t *Tag) Mouse(act text.Editor, e interface{}) {
	win := act.(*win.Win)
	if act := win; true {
		org := act.Origin()
		switch e := e.(type) {
		case mus.SnarfEvent:
			snarf(act)
		case mus.InsertEvent:
			paste(act)
		case mus.MarkEvent:
			if e.Button != 1 {
				t.r0, t.r1 = act.Dot()
			}
			q0 := org + act.IndexOf(p(e.Event))
			q1 := q0
			act.Sq = q0
			if e.Button == 1 && e.Double {
				q0, q1 = find.FreeExpand(act, q0)
				t.escR = image.Rect(-3, -3, 3, 3).Add(pt(e.Event))
			}
			act.Select(q0, q1)
		case mus.SweepEvent:
			if t.escR != image.ZR {
				if pt(e.Event).In(t.escR) {
					break
				}
				t.escR = image.ZR
				act.Select(act.Sq, act.Sq)
			}
			q0, q1 := act.Dot()
			//r0 := org+act.IndexOf(p(e.Event))
			sweeper := text.Sweeper(act)
			if act == t.Win {
				sweeper = mus.NewNopScroller(act)
			}
			act.Sq, q0, q1 = mus.Sweep(sweeper, e, 15, act.Sq, q0, q1, nil) //TODO (nil was act)
			if e.Button == 1 {
				act.Select(q0, q1)
			} else {
				act.Select(q0, q1)
			}
		case mus.SelectEvent:
			q0, q1 := act.Dot()
			if e.Button == 1 {
				act.Select(q0, q1)
				break
			}
			if e.Button == 2 || e.Button == 3 {
				q0, q1 := act.Dot()
				if q0 == q1 && text.Region3(q0, t.r0-1, t.r1) == 0 {
					// just use the existing selection and look
					q0, q1 = t.r0, t.r1
					act.Select(q0, q1)
				}
				if q0 == q1 {
					q0, q1 = find.ExpandFile(act.Bytes(), q0)
				}

				from := text.Editor(act)
				if from == t.Win {
					from = t
				}
				if e.Button == 3 {
					act.Select(q0, q1)
					act.Ctl() <- event.Look{
						Rec: event.Rec{
							Q0: q0,
							Q1: q1,
							P:  act.Bytes()[q0:q1],
						},
						From:    from,
						To:      []event.Editor{t.Body},
						Basedir: t.basedir,
						Name:    t.FileName(),
					}
				} else {
					act.Ctl() <- event.Cmd{
						Rec: event.Rec{
							Q0: q0, Q1: q1,
							P: act.Bytes()[q0:q1],
						},
						From:    from,
						To:      []event.Editor{t.Body},
						Basedir: t.basedir,
						Name:    t.FileName(),
					}
				}
			}
		}
	}
}

// Put
func (t *Tag) Handle(act text.Editor, e interface{}) {
	switch e := e.(type) {
	case mus.MarkEvent, mus.SweepEvent, mus.SelectEvent, mus.SnarfEvent, mus.InsertEvent:
		t.Mouse(act, e)
	case string:
		if e == "Redo" {
			//			act.Redo()
		} else if e == "Undo" {
			/*
				ev, err := t.Log.ReadAt(t.Log.Len()-1-t.offset)
				t.offset++
				if err != nil{
					t.SendFirst(err)
					return
				}
				ev2 := event.Invert(ev)
				switch ev2 := ev2.(type){
				case *event.Insert:
				t.Send(fmt.Errorf("INsert %#v\n", ev))
					act.Insert(ev2.P, ev2.Q0)
				case *event.Delete:
					q0,q1 := ev2.Q0, ev2.Q1
					if q0 > q1{
						q0,q1=q1,q0
					}
					if q0 != q1{
						q1--
					}
				t.Send(fmt.Errorf("Delete %#v\n", ev))
					act.Delete(q0,q1)
				}
				t.Send(fmt.Errorf("%#v\n", ev))
			*/
			//			act.Undo()
		} else if e == "Put" {
			t.Put()
		} else if e == "Get" {
			t.Get(t.FileName())
		}
		t.Mark()
	case *edit.Command:
		if e == nil {
			break
		}
		fn := e.Func()
		if fn != nil {
			fn(t.Body) // Always execute on body for now
		}
		t.Mark()
	case key.Event:
		if e.Direction == 2 {
			break
		}
		if e.Code == key.CodeI && e.Modifiers == key.ModControl {
			runGoImports(t, e)
			return
		}
		switch e.Code {
		case key.CodeEqualSign, key.CodeHyphenMinus:
			if e.Modifiers == key.ModControl {
				size := t.Body.Frame.Face.Height()
				if key.CodeHyphenMinus == e.Code {
					size -= 1
				} else {
					size += 1
				}
				if size < 3 {
					size = 6
				}
				//t.SetFont(t.Body.Frame.Font.NewSize(size))
				return
			}
		}
		ntab := int64(-1)
		if (e.Rune == '\n' || e.Rune == '\r') && act == t.Body {
			q0, q1 := act.Dot()
			if q0 == q1 {
				p := act.Bytes()
				l0, _ := find.Findlinerev(p, q0, 0)
				ntab = find.Accept(p, l0, []byte{'\t'})
				ntab -= l0 + 1
			}
		}
		kbd.SendClient(act, e)
		for ntab >= 0 {
			e.Rune = '\t'
			kbd.SendClient(act, e)
			ntab--
		}
		t.Mark()
	}
	t.dirty = true
}

func (t *Tag) Upload(wind screen.Window) {
	var wg sync.WaitGroup
	defer wg.Wait()
	if t.Body != nil && t.Body.Dirty() {
		wg.Add(1)
		go func() {
			t.Body.Upload()
			wg.Done()
		}()
	}
	if t.Win.Dirty() {
		wg.Add(1)
		go func() {
			t.Win.Upload()
			wg.Done()
		}()
	}
}

func (t *Tag) Refresh() {
	var wg sync.WaitGroup
	defer wg.Wait()
	if t.Body != nil {
		wg.Add(1)
		go func() {
			t.Body.Refresh()
			wg.Done()
		}()
	}
	if t.Win.Dirty() {
		wg.Add(1)
		go func() {
			t.Win.Refresh()
			wg.Done()
		}()
	}
}

func isdir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		if err == os.ErrNotExist {
			return false
		}
		fmt.Println(err)
		return false
	}
	return fi.IsDir()
}
