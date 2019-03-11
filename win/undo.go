package win

import "fmt"

// EnableUndoExperiment toggles the state of the Undo/Redo
// features. This is set to off by default. Eventually the variable
// will be removed.
//
// See-also:
// 	ins.go:/EnableUndoExperiment/
// 	del.go:/EnableUndoExperiment/
//
var EnableUndoExperiment = false

type Ops struct {
	q  int
	Op []Op
}

func (o *Ops) Insert(p []byte, q0 int64) int {
	println(fmt.Sprintf("#%d i,%q,\n", q0, p))
	return o.commit(OpIns{q0: q0, q1: int64(len(p)) + q0, p: p})
}
func (o *Ops) Delete(q0, q1 int64, p []byte) int {
	println(fmt.Sprintf("#%d,#%d d\n", q0, q1))
	return o.commit(OpDel{q0: q0, q1: q1, p: []byte(string(p[q0:q1]))})
}
func (o *Ops) Redo(w *Win) bool {
	if o.q == len(o.Op) {
		return false
	}
	o.Op[o.q].Do(w)
	o.q++
	return true
}
func (o *Ops) Undo(w *Win) bool {
	if o.q == 0 {
		return false
	}
	o.q--
	o.Op[o.q].Un().Do(w)
	return true
}
func (o *Ops) commit(op Op) int {
	if o.q != len(o.Op) {
		o.Op = append([]Op{}, o.Op[:o.q]...)
	}
	o.Op = append(o.Op, op)
	o.q++
	return 0
}

type (
	Op interface {
		Do(w *Win) int
		Un() Op
	}
	 OpIns op
	 OpDel op
	 op struct {
		q0, q1 int64
		p      []byte
	}
)

func (o OpIns) Do(w *Win) int { return w.insert(o.p, o.q0) }
func (o OpDel) Do(w *Win) int { return w.delete(o.q0, o.q1) }
func (o OpIns) Un() Op        { return OpDel(o) }
func (o OpDel) Un() Op        { return OpIns(o) }
