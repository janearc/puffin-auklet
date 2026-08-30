// package canvas is a cell buffer with alpha-style compositing.
//
// terminal sprites cannot be overlaid by concatenating styled strings: the
// escape sequences carry no geometry, so there is nothing to clip against.
// composing has to happen in cell space, and the string is produced once at
// the end.
package canvas

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Cell is one terminal cell. a nil FG or BG means "leave whatever is behind" --
// that is what makes a cutout possible. R == 0 means the cell is fully
// transparent and Blit skips it entirely.
type Cell struct {
	R  rune
	FG lipgloss.TerminalColor
	BG lipgloss.TerminalColor
}

func (c Cell) Empty() bool { return c.R == 0 }

type Canvas struct {
	w, h  int
	cells [][]Cell
}

func New(w, h int) *Canvas {
	c := &Canvas{w: w, h: h, cells: make([][]Cell, h)}
	for y := range c.cells {
		c.cells[y] = make([]Cell, w)
		for x := range c.cells[y] {
			c.cells[y][x] = Cell{R: ' '}
		}
	}
	return c
}

func (c *Canvas) Size() (int, int) { return c.w, c.h }

func (c *Canvas) Set(x, y int, cell Cell) {
	if x < 0 || y < 0 || x >= c.w || y >= c.h {
		return
	}
	c.cells[y][x] = cell
}

func (c *Canvas) Fill(cell Cell) {
	for y := 0; y < c.h; y++ {
		for x := 0; x < c.w; x++ {
			c.cells[y][x] = cell
		}
	}
}

// Text writes s at x,y. it does not wrap; anything past the edge is dropped.
func (c *Canvas) Text(x, y int, s string, fg, bg lipgloss.TerminalColor) {
	for i, r := range []rune(s) {
		c.Set(x+i, y, Cell{R: r, FG: fg, BG: bg})
	}
}

// Blit composites src at x,y, clipping at every edge. empty source cells are
// skipped, and a source cell with a nil background keeps the destination's --
// so a half-block whose lower half is transparent shows the canvas through it.
func (c *Canvas) Blit(src [][]Cell, x, y int) {
	for sy := range src {
		dy := y + sy
		if dy < 0 || dy >= c.h {
			continue
		}
		for sx := range src[sy] {
			dx := x + sx
			if dx < 0 || dx >= c.w {
				continue
			}
			s := src[sy][sx]
			if s.Empty() {
				continue
			}
			d := c.cells[dy][dx]
			out := Cell{R: s.R, FG: s.FG, BG: s.BG}
			if out.FG == nil {
				out.FG = d.FG
			}
			if out.BG == nil {
				out.BG = d.BG
			}
			c.cells[dy][dx] = out
		}
	}
}

func (c *Canvas) String() string { return Render(c.cells) }

// Render turns a cell grid into a printable string. adjacent cells sharing a
// style are emitted under one escape sequence, which matters: styling every
// cell separately makes a full-screen redraw roughly four times the bytes.
func Render(rows [][]Cell) string {
	var b strings.Builder
	for y, row := range rows {
		var run strings.Builder
		var curFG, curBG lipgloss.TerminalColor
		var have bool

		flush := func() {
			if run.Len() == 0 {
				return
			}
			s := lipgloss.NewStyle()
			if curFG != nil {
				s = s.Foreground(curFG)
			}
			if curBG != nil {
				s = s.Background(curBG)
			}
			b.WriteString(s.Render(run.String()))
			run.Reset()
		}

		for _, cell := range row {
			r := cell.R
			if r == 0 {
				r = ' '
			}
			if !have || cell.FG != curFG || cell.BG != curBG {
				flush()
				curFG, curBG, have = cell.FG, cell.BG, true
			}
			run.WriteRune(r)
		}
		flush()
		if y < len(rows)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// Fill2 fills a single row. useful for banded backdrops.
func (c *Canvas) Fill2(y int, cell Cell) {
	if y < 0 || y >= c.h {
		return
	}
	for x := 0; x < c.w; x++ {
		c.cells[y][x] = cell
	}
}

// Cells exposes the buffer for inspection and testing.
func (c *Canvas) Cells() [][]Cell { return c.cells }
