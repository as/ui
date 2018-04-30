package tag

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"

	"github.com/as/path"
	"github.com/as/text/find"
)

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

func (t *Tag) Open(basepath, title string) {
	t.basedir = path.DirOf(basepath)
	println(title)
	t.Get(title)
}

func (t *Tag) Dir() string {
	x := path.DirOf(t.FileName())
	if IsAbs(x) {
		return x
	}
	return filepath.Join(t.basedir, x)
}

func (t *Tag) abs() string {
	name := t.FileName()
	if !IsAbs(name) {
		name = filepath.Join(t.basedir, name)
	}
	return name
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
