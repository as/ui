package tag

import (
	"image"
	"log"
	"os"
	"sync"

	"github.com/as/srv/fs"
	"github.com/as/ui"

	"github.com/as/edit"
	"github.com/as/font"
	"github.com/as/frame"
	//	"github.com/as/ui/img"
	"github.com/as/ui/win"
	//"github.com/as/worm"
	"github.com/as/shiny/screen"
)

var (
	DefaultLabelHeight = 11
	DefaultConfig      = Config{
		Image:      true,
		FaceHeight: DefaultLabelHeight,
		Margin:     win.DefaultConfig.Margin,
		Facer:      font.NewFace,
		Color: [3]frame.Color{
			0: frame.ATag1,
			1: frame.A,
		},
	}
)

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

	ctl    chan<- interface{}
	Config Config
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
		Config:  *conf,
	}
	return t
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

func (t *Tag) Dirty() bool {
	return t.dirty || t.Win.Dirty() || (t.Body != nil && t.Body.Dirty())
}

func (t *Tag) Mark() {
	t.dirty = true
	t.Win.Mark()
}

//var crimson = image.NewUniform(color.RGBA{70, 40, 56, 255})

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

func mustCompile(prog string) *edit.Command {
	p, err := edit.Compile(prog)
	if err != nil {
		log.Printf("tag.go:/mustCompile/: failed to compile %q\n", prog)
		return nil
	}
	return p
}
