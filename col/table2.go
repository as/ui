package col

import (
	"image"

	"github.com/as/font"
	"github.com/as/ui"
	"github.com/as/ui/tag"
)

type Table2 struct {
	dev  ui.Dev
	ft   font.Face
	sp   image.Point
	size image.Point
	tdy  int

	Config *tag.Config
	Table
}

func NewTable2(dev ui.Dev, conf *tag.Config) Table2 {
	T := tag.New(dev, conf)
	T.Win.InsertString("New Delcol Sort	|", 0)

	return Table2{
		dev:    dev,
		ft:     conf.Facer(conf.FaceHeight),
		tdy:    conf.TagHeight(),
		Table:  Table{Tag: T},
		Config: T.Config,
	}
}

func (co *Table2) Loc() image.Rectangle {
	if co == nil {
		return image.ZR
	}
	return image.Rectangle{co.sp, co.sp.Add(co.size)}
}

func (co *Table2) Move(sp image.Point) {
	delta := sp.Sub(co.sp)
	co.Tag.Move(co.Tag.Loc().Min.Add(delta))
	for _, t := range co.List {
		t.Move(t.Loc().Min.Add(delta))
	}
	co.sp = sp
}

func (c *Table2) Dev() ui.Dev                { return c.dev }
func (c *Table2) Face() font.Face            { return c.ft }
func (c *Table2) ForceSize(size image.Point) { c.size = size }
