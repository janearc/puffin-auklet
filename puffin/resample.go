package puffin

import (
	"math"

	"github.com/charmbracelet/lipgloss"

	"github.com/janearc/puffin-auklet/canvas"
)

// roleWeight biases the vote when one subcell has to stand for several source
// pixels. a plain majority is the wrong rule for this picture: shrink it and
// the eye is outvoted by the face around it, the beak's ridge stripe is
// outvoted by the beak, and what is left is a blob with a wedge on it.
//
// so the roles that carry the bird's identity shout louder than the roles that
// are merely large. this is the same judgement the art itself encodes -- keep
// the cap, the face, the banded beak, the feet -- applied to sampling.
var roleWeight = map[byte]float64{
	'E': 7, // pupil
	'X': 6, // orbital ring
	'Y': 5, // the ridge stripe: one pixel wide in places, and load-bearing
	'O': 3, // feet
	'R': 2, // beak tip
	'B': 2, // beak base
	'D': 1.5,
	'W': 1,
	'K': 1,
	'V': 0.6, // wing tone is decoration; first to go
	'.': 1,
}

// SourceSize reports the side view's dimensions in source pixels.
func SourceSize() (w, h int) { return SideView.Size() }

// ColsFor returns the column count that preserves the bird's proportions at a
// given row count.
//
// a terminal cell is about twice as tall as it is wide, so a picture that is
// square on screen is not square in cells. this is the only place that ratio is
// applied; everything else works in whatever grid it is handed.
func ColsFor(rows int) int { return SideView.ColsFor(rows) }

func colsFor(sw, sh, rows int) int {
	return int(math.Round(float64(rows) * 2 * float64(sw) / float64(sh)))
}

// RowsToFit returns the largest row count whose sprite fits inside maxCols by
// maxRows, at least 3.
func RowsToFit(maxCols, maxRows int) int { return SideView.RowsToFit(maxCols, maxRows) }

func (s Sprite) RowsToFit(maxCols, maxRows int) int {
	rows := maxRows
	for rows > 3 && s.ColsFor(rows) > maxCols {
		rows--
	}
	if rows < 3 {
		rows = 3
	}
	return rows
}

// CellsAt renders the stencil into exactly cols by rows cells using gs.
//
// the pipeline is: resample the stencil into the subcell grid the glyph set
// gives us, then reduce each cell to the two colours a terminal cell can hold
// and pick the glyph whose lit subcells match that split. because every glyph
// set here is complete, the shape is always exact and only the colour is
// approximated.
func CellsAt(t Theme, gs GlyphSet, cols, rows int) [][]canvas.Cell {
	return SideView.CellsAt(t, gs, cols, rows, nil)
}

// CellsPosed is CellsAt with a pose stamped in first.
func CellsPosed(t Theme, gs GlyphSet, cols, rows int, pose Pose) [][]canvas.Cell {
	return SideView.CellsAt(t, gs, cols, rows, pose)
}

// CellsAt renders this sprite into exactly cols by rows cells.
func (s Sprite) CellsAt(t Theme, gs GlyphSet, cols, rows int, pose Pose) [][]canvas.Cell {
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	src := pose.applyTo(s.art)
	cw, ch := gs.Dims()
	subW, subH := cols*cw, rows*ch
	srcW, srcH := len(src[0]), len(src)

	// resample: each subcell takes the loudest role in the source rectangle
	// it covers. at scale 1 with half blocks this is an exact 1:1 copy.
	sub := make([][]byte, subH)
	for sy := 0; sy < subH; sy++ {
		sub[sy] = make([]byte, subW)
		y0, y1 := sy*srcH/subH, (sy+1)*srcH/subH
		if y1 <= y0 {
			y1 = y0 + 1
		}
		for sx := 0; sx < subW; sx++ {
			x0, x1 := sx*srcW/subW, (sx+1)*srcW/subW
			if x1 <= x0 {
				x1 = x0 + 1
			}
			sub[sy][sx] = loudest(src, x0, y0, x1, y1)
		}
	}

	out := make([][]canvas.Cell, rows)
	for cy := 0; cy < rows; cy++ {
		out[cy] = make([]canvas.Cell, cols)
		for cx := 0; cx < cols; cx++ {
			out[cy][cx] = pack(t, gs, sub, cx*cw, cy*ch)
		}
	}
	return out
}

func loudest(src []string, x0, y0, x1, y1 int) byte {
	var best byte
	var bestW float64 = -1
	seen := map[byte]float64{}
	order := make([]byte, 0, 8)
	for y := y0; y < y1 && y < len(src); y++ {
		for x := x0; x < x1 && x < len(src[y]); x++ {
			r := src[y][x]
			if _, ok := seen[r]; !ok {
				order = append(order, r)
			}
			seen[r] += roleWeight[r]
		}
	}
	// iterate in first-seen order so ties resolve the same way every run
	for _, r := range order {
		if seen[r] > bestW {
			best, bestW = r, seen[r]
		}
	}
	if bestW < 0 {
		return '.'
	}
	return best
}

// pack reduces one cell's subcells to two colours and the glyph that splits
// them. transparency is a third state: a subcell whose colour is nil is left
// unlit and unpainted, which is what lets a cutout keep its edge.
func pack(t Theme, gs GlyphSet, sub [][]byte, x0, y0 int) canvas.Cell {
	cw, ch := gs.Dims()
	n := cw * ch

	cols := make([]lipgloss.TerminalColor, n)
	weight := map[lipgloss.TerminalColor]float64{}
	order := make([]lipgloss.TerminalColor, 0, n)
	for i := 0; i < n; i++ {
		sy, sx := y0+i/cw, x0+i%cw
		role := sub[sy][sx]
		c := t.colorFor(role)
		cols[i] = c
		if _, ok := weight[c]; !ok {
			order = append(order, c)
		}
		weight[c] += roleWeight[role]
	}

	fg, bg := topTwo(order, weight)

	// the lit half of the cell must be the one carrying a colour, or the glyph
	// paints nothing.
	if fg == nil && bg != nil {
		fg, bg = bg, fg
	}
	if fg == nil {
		return canvas.Cell{}
	}

	mask := 0
	for i, c := range cols {
		if c == nil {
			continue // transparent subcells stay unlit
		}
		if c == fg || (c != bg && nearer(c, fg, bg)) {
			mask |= 1 << i
		}
	}
	if mask == 0 {
		if bg == nil {
			return canvas.Cell{}
		}
		return canvas.Cell{R: ' ', BG: bg}
	}
	return canvas.Cell{R: gs.Rune(mask), FG: fg, BG: bg}
}

func topTwo(order []lipgloss.TerminalColor, w map[lipgloss.TerminalColor]float64) (a, b lipgloss.TerminalColor) {
	var wa, wb float64 = -1, -1
	for _, c := range order {
		switch {
		case w[c] > wa:
			a, wa, b, wb = c, w[c], a, wa
		case w[c] > wb:
			b, wb = c, w[c]
		}
	}
	return a, b
}

// nearer reports whether c is closer to a than to b. an unreadable colour (a
// palette index, say) cannot be measured, so it falls to a.
func nearer(c, a, b lipgloss.TerminalColor) bool {
	if b == nil {
		return true
	}
	da, oka := dist2(c, a)
	db, okb := dist2(c, b)
	if !oka || !okb {
		return true
	}
	return da <= db
}

func dist2(x, y lipgloss.TerminalColor) (float64, bool) {
	xs, ok1 := resolve(x)
	ys, ok2 := resolve(y)
	if !ok1 || !ok2 || len(xs) == 0 || len(ys) == 0 {
		return 0, false
	}
	return dist(xs[len(xs)-1], ys[len(ys)-1]), true
}
