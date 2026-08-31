package scene

import (
	"testing"

	"github.com/janearc/puffin-auklet/auklet"
	"github.com/janearc/puffin-auklet/canvas"
	"github.com/janearc/puffin-auklet/themes"
)

func base() Opts {
	win := Window{W: 40, H: 20, ScreenX: 3, ScreenY: 2}
	rows := 12
	return Opts{
		Theme: themes.All[0], Glyphs: auklet.Quadrant, Rows: rows,
		Win: win, Cutout: true, W: 60, H: 26,
		SpriteX: win.WorldX + (win.W-auklet.ColsFor(rows))/2,
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
	cols := auklet.ColsFor(b.Rows)

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
	if o.SpriteX != before.SpriteX || o.SpriteY != before.SpriteY || o.Win != before.Win {
		t.Error("Build mutated its options")
	}

	// walk the window to the bird rather than moving the bird back
	o.Win.WorldX = o.SpriteX - (o.Win.W-auklet.ColsFor(o.Rows))/2
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

// Backdrop is drawable on its own, into any canvas, by name.
//
// It was private and Build was the only way to reach it -- but Build composes
// a workbench screen, desktop and window box included. A caller that wants a
// strip of fire along the bottom of its own interface had to take the
// picture frame with it.
func TestBackdropDrawsIntoAnyCanvas(t *testing.T) {
	th := themes.All[0].Theme
	field := th.Background

	// a wide, short strip -- the shape a caller wants along the bottom of a
	// terminal, and nothing like the workbench's window
	c := canvas.New(120, 6)
	if !Backdrop(c, "fireplace", th, field, 0) {
		t.Fatal("fireplace is not a known backdrop")
	}
	w, h := c.Size()
	if w != 120 || h != 6 {
		t.Fatalf("the canvas changed size: %dx%d", w, h)
	}
	// the fire is at the BOTTOM: embers on the last row, and the top row of
	// a six-row strip is above the flames
	ink := func(y int) int {
		n := 0
		for x := 0; x < w; x++ {
			if cellAt(c, x, y).R != ' ' && cellAt(c, x, y).R != 0 {
				n++
			}
		}
		return n
	}
	if ink(h-1) == 0 {
		t.Error("no embers on the bottom row")
	}
	if ink(h-1) < ink(0) {
		t.Error("the fire is upside down: more ink at the top than at the coals")
	}
}

// Every name in the roster draws something, and an unknown one draws a field
// rather than failing: a wrong name should look like nothing, never like a
// crash.
func TestEveryBackdropNameIsDrawable(t *testing.T) {
	th := themes.All[0].Theme
	for _, name := range BackdropNames {
		c := canvas.New(40, 10)
		if !Backdrop(c, name, th, th.Background, 3) {
			t.Errorf("%s is in BackdropNames and is not drawable", name)
		}
	}
	c := canvas.New(40, 10)
	if Backdrop(c, "no-such-backdrop", th, th.Background, 0) {
		t.Error("an unknown name reported itself as known")
	}
	if w, h := c.Size(); w != 40 || h != 10 {
		t.Fatalf("an unknown name damaged the canvas: %dx%d", w, h)
	}
}

// The animated ones move with Frame, and the static ones do not -- which is
// what lets a caller pass a frame counter unconditionally.
func TestFrameMovesOnlyTheAnimatedBackdrops(t *testing.T) {
	th := themes.All[0].Theme
	render := func(name string, frame int) string {
		c := canvas.New(60, 8)
		Backdrop(c, name, th, th.Background, frame)
		return c.String()
	}
	for _, name := range []string{"lights", "fireplace"} {
		if render(name, 0) == render(name, 7) {
			t.Errorf("%s did not move between frames", name)
		}
	}
	for _, name := range []string{"flat", "grid", "text", "bands"} {
		if render(name, 0) != render(name, 7) {
			t.Errorf("%s moved, and it is not an animated backdrop", name)
		}
	}
}

// Build still draws the same screen it always did: the index path is a thin
// wrapper now, and the workbench must not have noticed.
func TestBuildStillComposesThroughTheIndex(t *testing.T) {
	o := base()
	for i := range BackdropNames {
		o.Backdrop = i
		if Build(o) == nil {
			t.Fatalf("backdrop %d (%s) built nothing", i, BackdropNames[i])
		}
	}
}
