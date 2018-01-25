package win

func (w *Win) Fill() {
	if w.Frame.Full() {
		return
	}
	for !w.Frame.Full() {
		qep := w.org + w.Nchars
		n := max(0, min(w.Len()-qep, 2000))
		if n == 0 {
			break
		}
		rp := w.Bytes()[qep : qep+n]
		m := len(rp)
		nl := w.MaxLine() - w.Line()
		m = 0
		i := int64(0)
		for i < n {
			if rp[i] == '\n' {
				m++
				if m >= nl {
					i++
					break
				}
			}
			i++
		}
		w.Frame.Insert(rp[:i], w.Nchars)
	}
	w.Flush()
}
