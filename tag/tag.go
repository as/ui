package tag

import (
	"bufio"
	"bytes"
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/as/srv/fs"
	"github.com/as/ui"

	"github.com/as/edit"
	"github.com/as/font"
	"github.com/as/frame"
	"github.com/as/path"
	"github.com/as/text/find"
	//	"github.com/as/ui/img"
	"github.com/as/ui/win"
	//"github.com/as/worm"
	"github.com/as/shiny/screen"
	"golang.org/x/mobile/event/mouse"
)

var (
	DefaultFaceHeight = win.DefaultFaceHeight
	DefaultMargin     = win.DefaultMargin
	DefaultConfig     = Config{
		Image:      true,
		FaceHeight: DefaultFaceHeight,
		Margin:     DefaultMargin,
		Facer:      font.NewFace,
		Color: [3]frame.Color{
			0: frame.ATag1,
			1: frame.A,
		},
	}
)

func p(e mouse.Event) image.Point {
	return image.Pt(int(e.X), int(e.Y))
}

type Tag struct {
	Vis
	sp   image.Point
	size image.Point

	*win.Win
	Body      Window //*win.Win
	Scrolling bool
	fs.Fs

	basedir string
	dirty   bool
	r0, r1  int64
	escR    image.Rectangle

	ctl    chan<- interface{}
	Config *Config
}

func New(dev ui.Dev, conf *Config) *Tag {
	conf = validConfig(conf)
	wd, _ := os.Getwd() // TODO(as): BUG!!!
	t := &Tag{
		basedir: wd,
		Fs:      conf.Filesystem,
		Win:     win.New(dev, conf.TagConfig()),
		Body:    win.New(dev, conf.WinConfig()),
		ctl:     conf.Ctl,
		Config:  conf,
	}
	return t
}

func (w *Tag) SetFont(ft font.Face) {
	body := w.Body.(*win.Win)
	if body == nil {
		return
	}
	if ft.Height() < 3 || w.Body == nil {
		return
	}
	body.SetFont(ft)
	w.dirty = true
	w.Mark()
	body.Refresh()
}

func (t *Tag) Dirty() bool {
	return t.dirty || t.Win.Dirty() || (t.Body != nil && t.Body.Dirty())
}

func (t *Tag) Mark() {
	t.dirty = true
	t.Win.Mark()
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

func pt(e mouse.Event) image.Point {
	return image.Pt(int(e.X), int(e.Y))
}

var (
	crimson = image.NewUniform(color.RGBA{70, 40, 56, 255})
)

func (t *Tag) Upload(wind screen.Window) {
	var wg sync.WaitGroup
	defer wg.Wait()
	if t.Body != nil && t.Body.Dirty() {
		wg.Add(1)
		go func() {
			{
				//	dst := t.Win.Buffer().RGBA()
				//	src := crimson
				//	draw.Draw(dst, image.Rect(0, 0, 8, 8).Add(t.Win.Bounds().Min), src, image.ZP, draw.Src)
			}
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
