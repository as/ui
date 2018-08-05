package tag

import (
	"image"
	"log"

	"github.com/as/srv/fs"
	"github.com/as/ui"

	"github.com/as/edit"
	"github.com/as/font"
	"github.com/as/frame"
	"github.com/as/ui/img"
	"github.com/as/ui/win"
	//"github.com/as/worm"
)

var DefaultLabelHeight = 11

var DefaultConfig = Config{
	FaceHeight: DefaultLabelHeight,
	Margin:     win.DefaultConfig.Margin,
	Facer:      font.NewFace,
	Color: [3]frame.Color{
		frame.Theme(image.Black, image.White, image.White, image.Black),
		frame.Theme(image.White, image.Black, image.Black, image.White),
	},
}

type Tag struct {
	Vis
	sp   image.Point
	size image.Point

	Win *win.Win
	Window //*win.Win
	Scrolling bool
	fs.Fs

	dirty bool

	ctl    chan<- interface{}
	Config Config
}

func New(dev ui.Dev, conf *Config) *Tag {
	conf = validConfig(conf)
	if conf.Image {
		return &Tag{
			Fs:     conf.Filesystem,
			Win:    win.New(dev, conf.TagConfig()),
			Body:   img.New(dev, nil),
			ctl:    conf.Ctl,
			Config: *conf,
		}
	}
	return &Tag{
		Fs:     conf.Filesystem,
		Win:    win.New(dev, conf.TagConfig()),
		Body:   win.New(dev, conf.WinConfig()),
		ctl:    conf.Ctl,
		Config: *conf,
	}
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

func (t *Tag) Upload() {
	t.Body.Upload()
	t.Win.Upload()
}

func (t *Tag) Refresh() {
	t.Body.Refresh()
	t.Win.Refresh()
}

func mustCompile(prog string) *edit.Command {
	p, err := edit.Compile(prog)
	if err != nil {
		log.Printf("tag.go:/mustCompile/: failed to compile %q\n", prog)
		return nil
	}
	return p
}
