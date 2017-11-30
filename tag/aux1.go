package tag

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/mobile/event/key"
)

func runGoImports(t *Tag, e key.Event) {
	if !strings.HasSuffix(t.FileName(), ".go") {
		t.Window().Send(fmt.Errorf("Wont run goimports, file is %q", t.FileName()))
		return
	}
	cmd := exec.Command("goimports")
	cmd.Stdin = bytes.NewReader(t.Body.Bytes())
	b := new(bytes.Buffer)
	cmd.Stdout = b
	cmd.Run()
	if b.Len() < len("package") {
		t.Window().Send(fmt.Errorf("goimports failed", t.FileName()))
	}
	q0, q1 := t.Body.Dot()
	t.Body.Delete(0, t.Body.Len())
	t.Body.Insert(b.Bytes(), 0)
	t.Mark()
	t.Select(q0, q1)
}
