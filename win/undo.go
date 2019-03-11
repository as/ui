package win

// EnableUndoExperiment toggles the state of the Undo/Redo
// features. This is set to off by default. Eventually the variable
// will be removed.
//
var EnableUndoExperiment = false

type Op interface {
	Do(w *Win) int
	Un() Op
}
type op struct {
	q0, q1 int64
	p      []byte
}
type OpIns op
type OpDel op
