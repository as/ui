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

	"github.com/as/edit"
	"github.com/as/frame"
	"github.com/as/path"
	"github.com/as/shrew"
	"github.com/as/text"
	"github.com/as/text/action"
	"github.com/as/text/find"
	"github.com/as/ui/win"
	"golang.org/x/image/font"
	//"github.com/as/worm"
)

// Put
var (
	Buttonsdown = 0
	noselect    bool
	lastclickpt image.Point
)

type Tag struct {
	*win.Win
	Body      *win.Win
	sp        image.Point
	Scrolling bool
	scrolldy  int
	dirty     bool
	r0, r1    int64
	escR      image.Rectangle
	offset    int64
	basedir   string
}

func (w *Tag) SetFont(ft font.Face) {
	if frame.Dy(ft) < 3 || w.Body == nil {
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
	r := t.Win.Bounds()
	if t.Body != nil {
		r.Max.Y += t.Body.Bounds().Dy()
	}
	return r
}

// TagSize returns the size of a tag given the font
func TagSize(ft font.Face) int {
	dy := frame.Dy(ft)
	return dy + dy/2
}

// TagPad returns the padding for the tag given the window's padding
// always returns an x-aligned point
func TagPad(wpad image.Point) image.Point {
	return image.Pt(wpad.X, 3)
}

type Config struct {
	TagHeight int
	TagColor  *frame.Color
	FontFunc  func(int) font.Face
	FontSize  int
	Drawer    frame.Drawer
	Flag      int
	Pad       image.Point
	Color     *frame.Color
}

func (c *Config) check() {
	if c.FontFunc == nil {
		c.FontFunc = frame.NewGoMono
	}
	if c.FontSize == 0 {
		c.FontSize = 11
	}
	if c.TagHeight == 0 {
		c.TagHeight = c.FontSize + c.FontSize/2
	}
	if c.TagColor == nil {
		c.TagColor = &frame.ATag1
	}
	if c.Color == nil {
		c.Color = &frame.A
	}
	if c.Pad == image.ZP {
		c.Pad = image.Pt(15, 15)
	}
}

// Put
func New(c *shrew.Client, sp, size image.Point, conf *Config) *Tag {
	if conf == nil {
		conf = &Config{}
	}
	conf.check()
	// Make the main tag
	tagY := conf.TagHeight

	wtag := win.New(c, sp, image.Pt(size.X, tagY), &win.Config{
		Flag:   conf.Flag,
		Pad:    image.Pt(2, 2),
		Face:   conf.FontFunc(11),
		Drawer: conf.Drawer,
		Color:  conf.TagColor,
	})

	sp = sp.Add(image.Pt(0, tagY))
	size = size.Sub(image.Pt(0, tagY))
	if size.Y < tagY {
		return &Tag{sp: sp, Win: wtag, Body: nil}
	}

	// Make window
	//	cols.Back = Yellow
	//	ft = font.Clone(ft, ft.Size())
	//	ft.SetLetting(ft.Size() / 3)
	w := win.New(c, sp, size, &win.Config{
		Flag:   conf.Flag,
		Pad:    conf.Pad,
		Face:   conf.FontFunc(conf.FontSize),
		Color:  conf.Color,
		Drawer: conf.Drawer,
	})

	wd, _ := os.Getwd()
	return &Tag{sp: sp, Win: wtag, Body: w, basedir: wd}
}

func (t *Tag) Move(pt image.Point) {
	/*
		t.Win.Move(pt)
		if t.Body == nil {
			return
		}
		pt.Y += t.Win.Loc().Dy()
		t.Body.Move(pt)
	*/
}

func (t *Tag) Resize(pt image.Point) {}

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

func (t *Tag) Close() (err error) { return nil }

func (t *Tag) Dir() string {
	x := path.DirOf(t.FileName())
	if filepath.IsAbs(x) {
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
	wtag.Insert([]byte(abs+"\tPut Del |"), 0)
	wtag.Refresh()
}
func (t *Tag) getbody(abs, addr string) {
	w := t.Body
	w.Delete(0, w.Len())
	w.Insert(t.readfile(abs), 0)
	w.Select(0, 0)
	w.SetOrigin(0, true)
	if addr != "" {
		//		w.SendFirst(mustCompile(addr))
	}
}

func (t *Tag) Get(name string) {
	w := t.Body
	if w == nil {
		//		w.SendFirst(fmt.Errorf("tag: window has no body for get request %q\n", name))
		return
	}
	if name == "" {
		t.fixtag("")
		return
	}

	abs := ""
	name, addr := action.SplitPath(name)
	if filepath.IsAbs(name) && path.Exists(name) {
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
	if !filepath.IsAbs(name) {
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
	//	t.Window().Send(fmt.Errorf("Put %q", name))
	writefile(name, t.Body.Bytes())
	return nil
}
func (t *Tag) Mouse(act text.Editor, e interface{}) {}

// Put
func (t *Tag) Handle(act text.Editor, e interface{}) {}

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
func isfile(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
