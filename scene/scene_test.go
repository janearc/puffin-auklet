package scene

import (
	"testing"

	"github.com/janearc/puffin-auklet/canvas"
	"github.com/janearc/puffin-auklet/puffin"
	"github.com/janearc/puffin-auklet/themes"
)

func base() Opts {
	win := Window{W: 40, H: 20, ScreenX: 3, ScreenY: 2}
	rows := 12
	return Opts{
		Theme: themes.All[0], Glyphs: puffin.Quadrant, Rows: rows,
		Win: win, Cutout: true, W: 60, H: 26,
		SpriteX: win.WorldX + (win.W-puffin.ColsFor(rows))/2,
		SpriteY: win.WorldY + (win.H-rows)/2,
	}
}

// how many cells inside the window's interior carry sprite colour rather than
// backdrop. counting the beak tip alone is enough: nothing else uses it.
func beakCells(o Opts) int {
	c := Build(o)
	tip := o.Theme.Theme.BeakTip
	n := 0
	for y := 0; y < o.Win.H; y++ {
		for x := 0; x < o.Win.W; x++ {
			cell := cellAt(c, o.Win.ScreenX+x, o.Win.ScreenY+y)
			if cell.FG == tip || cell.BG == tip {
				n++
			}
		}
	}
	return n
}

func cellAt(c *canvas.Canvas, x, y int) canvas.Cell {
	w, h := c.Size()
	if x < 0 || y < 0 || x >= w || y >= h {
		return canvas.Cell{}
	}
	return c.Cells()[y][x]
}

func TestVisibility(t *testing.T) {
	b := base()
	cols := puffin.ColsFor(b.Rows)

	cases := []struct {
		name string
		x, y int
		want bool
	}{
		{"centred", b.SpriteX, b.SpriteY, true},
		{"one column still showing", b.Win.WorldX - cols + 1, b.SpriteY, true},
		{"one column past the edge", b.Win.WorldX - cols, b.SpriteY, false},
		{"far left", -500, b.SpriteY, false},
		{"far below", b.SpriteX, 900, false},
	}
	for _, c := range cases {
		o := base()
		o.SpriteX, o.SpriteY = c.x, c.y
		if got := Visible(o); got != c.want {
			t.Errorf("%s: Visible = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestSpriteClipsAtTheWindowEdge(t *testing.T) {
	full := beakCells(base())
	if full == 0 {
		t.Fatal("centred sprite drew no beak at all")
	}

	// far enough left that the beak itself crosses the edge: the beak is at
	// the sprite's leading edge, so a small nudge clips nothing.
	partial := base()
	partial.SpriteX = partial.Win.WorldX - 6
	got := beakCells(partial)
	if got == 0 || got >= full {
		t.Errorf("half-off sprite drew %d beak cells, want between 1 and %d", got, full-1)
	}

	gone := base()
	gone.SpriteX = -500
	if got := beakCells(gone); got != 0 {
		t.Errorf("out-of-view sprite drew %d beak cells, want 0", got)
	}
}

// object permanence: leaving the window is not an event that happens TO the
// sprite. its coordinates are untouched, and bringing the view back to them
// brings the bird back whole.
func TestSpriteSurvivesLeavingTheView(t *testing.T) {
	full := beakCells(base())

	o := base()
	o.SpriteX, o.SpriteY = 4000, -4000
	if Visible(o) {
		t.Fatal("sprite at 4000,-4000 should not be visible")
	}
	if n := beakCells(o); n != 0 {
		t.Fatalf("invisible sprite still drew %d cells", n)
	}

	before := o
	Build(o)
	if o != before {
		t.Error("Build mutated its options")
	}

	// walk the window to the bird rather than moving the bird back
	o.Win.WorldX = o.SpriteX - (o.Win.W-puffin.ColsFor(o.Rows))/2
	o.Win.WorldY = o.SpriteY - (o.Win.H-o.Rows)/2
	if !Visible(o) {
		t.Fatal("window moved onto the sprite but it is still not visible")
	}
	if n := beakCells(o); n != full {
		t.Errorf("recovered sprite drew %d beak cells, want %d", n, full)
	}
}

func TestOutOfViewMarker(t *testing.T) {
	dirs := []struct {
		name string
		x, y int
		want rune
	}{
		{"left", -500, 0, '<'},
		{"right", 500, 0, '>'},
		{"above", 0, -500, '^'},
		{"below", 0, 500, 'v'},
	}
	for _, d := range dirs {
		o := base()
		if d.x != 0 {
			o.SpriteX = d.x
		}
		if d.y != 0 {
			o.SpriteY = d.y
		}
		c := Build(o)

		found := false
		w, h := c.Size()
		for y := 0; y < h && !found; y++ {
			for x := 0; x < w; x++ {
				if c.Cells()[y][x].R == d.want {
					found = true
					break
				}
			}
		}
		if !found {
			t.Errorf("%s: no %q marker on the border for an out-of-view sprite",
				d.name, string(d.want))
		}
	}
}
