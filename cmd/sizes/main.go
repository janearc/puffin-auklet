// sizes dumps the sprite alone at a given glyph set and row count, one line
// per cell, for offline inspection.
//
//	go run ./cmd/sizes <themeIdx> <glyphset 0=half 1=quad 2=sextant> <rows>
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/charmbracelet/lipgloss"

	"github.com/janearc/puffin-auklet/puffin"
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
	ti, gsi, rows := argi(0, 0), argi(1, 1), argi(2, 22)
	gs := puffin.GlyphSet(gsi)
	t := themes.All[ti%len(themes.All)].Theme
	cols := puffin.ColsFor(rows)

	grid := puffin.CellsAt(t, gs, cols, rows)
	cw, ch := gs.Dims()
	fmt.Printf("# %d %d %d %d\n", cols, rows, cw, ch)
	for y, row := range grid {
		for x, cell := range row {
			r := cell.R
			if r == 0 {
				r = ' '
			}
			fmt.Printf("%d %d %d %s %s\n", x, y, r, hex(cell.FG), hex(cell.BG))
		}
	}
}
