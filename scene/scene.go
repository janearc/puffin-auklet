// package scene composes the auklet into a window, and the window onto the
// screen. it is shared by the interactive viewer and the headless renderers so
// that what CI checks and what you look at are the same code path.
//
// three coordinate spaces, and keeping them apart is the whole design:
//
//	world   where the bird actually is. unbounded, never clamped.
//	window  a rectangle of world space that is currently being shown.
//	screen  where that window is drawn in the terminal.
//
// the bird's world position survives being scrolled out of view, moved off the
// window, or having the window walk away from it. it is not clipped, hidden, or
// snapped back to an edge -- it is simply somewhere the window is not looking.
package scene

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/janearc/puffin-auklet/canvas"
	"github.com/janearc/puffin-auklet/puffin"
	"github.com/janearc/puffin-auklet/themes"
)

// the backdrops exist to make transparency falsifiable. a cutout over a flat
// field is indistinguishable from an opaque sprite whose background happens to
// match; over text or bands it is obvious immediately.
var BackdropNames = []string{"flat", "grid", "text", "bands"}

// Window is a rectangle of world space, shown somewhere on screen.
type Window struct {
	WorldX, WorldY   int // world coordinate at the window's top-left
	W, H             int // interior size, in cells
	ScreenX, ScreenY int // where the interior's top-left lands on screen
}

type Opts struct {
	Theme    themes.Named
	Backdrop int
	Glyphs   puffin.GlyphSet
	Rows     int // sprite height in cells; width follows from the aspect

	SpriteX, SpriteY int // WORLD coordinates
	Win              Window
	Cutout           bool
	Pose             puffin.Pose

	W, H int // screen size
}

const filler = "dodo whoops auklet corvid vaporwave bladerunner stardew "

// Build returns the screen.
func Build(o Opts) *canvas.Canvas {
	t := o.Theme.Theme
	field := t.Background
	if o.Cutout {
		// the sprite must not carry a field of its own, or there is nothing to
		// see through.
		t.Background = nil
	}

	screen := canvas.New(o.W, o.H)
	desktop(screen, o.Theme.Theme)

	win := canvas.New(o.Win.W, o.Win.H)
	backdrop(win, o.Backdrop, t, field)

	cols := puffin.ColsFor(o.Rows)
	win.Blit(puffin.CellsPosed(t, o.Glyphs, cols, o.Rows, o.Pose),
		o.SpriteX-o.Win.WorldX, o.SpriteY-o.Win.WorldY)

	screen.BlitCanvas(win, o.Win.ScreenX, o.Win.ScreenY)
	screen.Box(o.Win.ScreenX, o.Win.ScreenY, o.Win.W, o.Win.H,
		o.Theme.Theme.Light, o.Theme.Theme.Background)

	if !Visible(o) {
		marker(screen, o)
	}
	return screen
}

// desktop is what surrounds the window: the theme's ground, stippled, so the
// window reads as a window rather than as the whole screen.
func desktop(c *canvas.Canvas, t puffin.Theme) {
	c.Fill(canvas.Cell{R: ' ', BG: t.Background})
	w, h := c.Size()
	for y := 0; y < h; y++ {
		for x := (y % 2) * 2; x < w; x += 4 {
			c.Set(x, y, canvas.Cell{R: '·', FG: t.Wing, BG: t.Background})
		}
	}
}

func backdrop(c *canvas.Canvas, which int, t puffin.Theme, field lipgloss.TerminalColor) {
	w, h := c.Size()
	c.Fill(canvas.Cell{R: ' ', BG: field})

	switch BackdropNames[which%len(BackdropNames)] {
	case "grid":
		for y := 0; y < h; y += 2 {
			for x := 0; x < w; x += 4 {
				c.Set(x, y, canvas.Cell{R: '·', FG: t.Wing, BG: field})
			}
		}
	case "text":
		line := strings.Repeat(filler, w/len(filler)+2)
		for y := 0; y < h; y++ {
			off := (y * 7) % len(filler)
			c.Text(0, y, line[off:off+w], t.Stripe, field)
		}
	case "bands":
		band := []lipgloss.TerminalColor{t.Wing, t.BeakBase, t.Stripe, field}
		for y := 0; y < h; y++ {
			c.FillRow(y, canvas.Cell{R: ' ', BG: band[(y/2)%len(band)]})
		}
	}
}

// SpriteRect returns the sprite's bounds in world space.
func SpriteRect(o Opts) (x, y, w, h int) {
	return o.SpriteX, o.SpriteY, puffin.ColsFor(o.Rows), o.Rows
}

// Visible reports whether any part of the sprite falls inside the window.
func Visible(o Opts) bool {
	x, y, w, h := SpriteRect(o)
	return x < o.Win.WorldX+o.Win.W && x+w > o.Win.WorldX &&
		y < o.Win.WorldY+o.Win.H && y+h > o.Win.WorldY
}

// marker puts a pointer on the window's border in the direction of a bird that
// has gone out of view. it is the visible half of object permanence: the sprite
// still has a position, and the window can tell you which way it is.
func marker(c *canvas.Canvas, o Opts) {
	x, y, w, h := SpriteRect(o)
	cx, cy := x+w/2, y+h/2
	midX := o.Win.WorldX + o.Win.W/2
	midY := o.Win.WorldY + o.Win.H/2

	fg := o.Theme.Theme.BeakTip
	bg := o.Theme.Theme.Background

	if cx < o.Win.WorldX || cx >= o.Win.WorldX+o.Win.W {
		r, sx := '>', o.Win.ScreenX+o.Win.W
		if cx < midX {
			r, sx = '<', o.Win.ScreenX-1
		}
		sy := o.Win.ScreenY + clamp(cy-o.Win.WorldY, 0, o.Win.H-1)
		c.Set(sx, sy, canvas.Cell{R: r, FG: fg, BG: bg})
	}
	if cy < o.Win.WorldY || cy >= o.Win.WorldY+o.Win.H {
		r, sy := 'v', o.Win.ScreenY+o.Win.H
		if cy < midY {
			r, sy = '^', o.Win.ScreenY-1
		}
		sx := o.Win.ScreenX + clamp(cx-o.Win.WorldX, 0, o.Win.W-1)
		c.Set(sx, sy, canvas.Cell{R: r, FG: fg, BG: bg})
	}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Demo builds a scene from positional command-line arguments. it exists so the
// headless renderers stay thin and agree with each other on what the arguments
// mean.
//
//	[theme] [backdrop] [rows] [cutout 0|1] [dx] [dy]
func Demo(args []string) (Opts, string, string) {
	argi := func(i, def int) int {
		if len(args) > i {
			if v, err := strconv.Atoi(args[i]); err == nil {
				return v
			}
		}
		return def
	}

	ti := argi(0, 0)
	cur := themes.All[((ti%len(themes.All))+len(themes.All))%len(themes.All)]

	const w, h = 72, 30
	win := Window{W: 46, H: 22, ScreenX: 3, ScreenY: 2}
	rows := argi(2, 14)
	o := Opts{
		Theme: cur, Backdrop: argi(1, 0), Glyphs: puffin.Quadrant, Rows: rows,
		Win: win, Cutout: argi(3, 0) == 1, W: w, H: h,
	}
	o.SpriteX = win.WorldX + (win.W-puffin.ColsFor(rows))/2 + argi(4, 0)
	o.SpriteY = win.WorldY + (win.H-rows)/2 + argi(5, 0)
	return o, cur.Name, cur.Note
}
