// dump prints a composed scene as one line per cell: "x y codepoint fg bg".
// the rune is a decimal code point, not a quoted character: a quoted space
// contains a space and silently shifts every field after it.
// it exists so a renderer can be verified without parsing escape sequences,
// which is a second, flakier implementation of the thing under test.
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/charmbracelet/lipgloss"

	"github.com/janearc/puffin-auklet/canvas"
	"github.com/janearc/puffin-auklet/puffin"
	"github.com/janearc/puffin-auklet/scene"
	"github.com/janearc/puffin-auklet/themes"
)

func argi(i, def int) int {
	if len(os.Args) > i+1 {
		if v, err := strconv.Atoi(os.Args[i+1]); err == nil {
			return v
		}
	}
	return def
}

func hex(c lipgloss.TerminalColor) string {
	if c == nil {
		return "-"
	}
	if v, ok := c.(lipgloss.Color); ok {
		return string(v)
	}
	return "?"
}

func main() {
	ti, bi, sc, cut := argi(0, 0), argi(1, 0), argi(2, 1), argi(3, 0)
	cur := themes.All[ti%len(themes.All)]

	w, h := 60, 26
	sw, sh := puffin.SizeAt(sc)
	c := scene.Build(scene.Opts{
		Theme: cur, Backdrop: bi, Scale: sc, Cutout: cut == 1,
		X: (w - sw) / 2, Y: (h - sh) / 2, W: w, H: h,
	})

	fmt.Printf("# %d %d\n", w, h)
	var rows [][]canvas.Cell = c.Cells()
	for y, row := range rows {
		for x, cell := range row {
			r := cell.R
			if r == 0 {
				r = ' '
			}
			fmt.Printf("%d %d %d %s %s\n", x, y, r, hex(cell.FG), hex(cell.BG))
		}
	}
}
