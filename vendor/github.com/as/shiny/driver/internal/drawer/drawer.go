// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package drawer provides functions that help implement screen.Drawer methods.
package drawer // import "github.com/as/shiny/driver/internal/drawer"

import (
	"image"
	"image/draw"

	"github.com/as/shiny/math/f64"
	"github.com/as/shiny/screen"
)

// Copy implements the Copy method of the screen.Drawer interface by calling
// the Draw method of that same interface.
func Copy(dst screen.Drawer, dp image.Point, src screen.Texture, sr image.Rectangle, op draw.Op, opts *screen.DrawOptions) {
	dst.Draw(f64.Aff3{
		1, 0, float64(dp.X - sr.Min.X),
		0, 1, float64(dp.Y - sr.Min.Y),
	}, src, sr, op, opts)
}

// Scale implements the Scale method of the screen.Drawer interface by calling
// the Draw method of that same interface.
func Scale(dst screen.Drawer, dr image.Rectangle, src screen.Texture, sr image.Rectangle, op draw.Op, opts *screen.DrawOptions) {
	rx := float64(dr.Dx()) / float64(sr.Dx())
	ry := float64(dr.Dy()) / float64(sr.Dy())
	dst.Draw(f64.Aff3{
		rx, 0, float64(dr.Min.X) - rx*float64(sr.Min.X),
		0, ry, float64(dr.Min.Y) - ry*float64(sr.Min.Y),
	}, src, sr, op, opts)
}
