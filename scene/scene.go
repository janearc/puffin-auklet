// package scene composes the puffin onto a backdrop. it is shared by the
// interactive viewer and the headless renderer so that what ci checks and what
// you look at are the same code path.
package scene

import (
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

type Opts struct {
	Theme    themes.Named
	Backdrop int
	Scale    int
	X, Y     int
	Cutout   bool
	W, H     int
}

const filler = "dodo whoops puffin corvid vaporwave bladerunner stardew "

func Build(o Opts) *canvas.Canvas {
	t := o.Theme.Theme
	field := t.Background
	if o.Cutout {
		// the sprite must not carry a field of its own, or there is nothing to
		// see through.
		t.Background = nil
	}

	c := canvas.New(o.W, o.H)
	c.Fill(canvas.Cell{R: ' ', BG: field})

	switch BackdropNames[o.Backdrop%len(BackdropNames)] {
	case "grid":
		for y := 0; y < o.H; y += 2 {
			for x := 0; x < o.W; x += 4 {
				c.Set(x, y, canvas.Cell{R: '·', FG: t.Wing, BG: field})
			}
		}
	case "text":
		line := strings.Repeat(filler, o.W/len(filler)+2)
		for y := 0; y < o.H; y++ {
			off := (y * 7) % len(filler)
			c.Text(0, y, line[off:off+o.W], t.Stripe, field)
		}
	case "bands":
		band := []lipgloss.TerminalColor{t.Wing, t.BeakBase, t.Stripe, field}
		for y := 0; y < o.H; y++ {
			c.Fill2(y, canvas.Cell{R: ' ', BG: band[(y/2)%len(band)]})
		}
	}

	c.Blit(puffin.Cells(t, o.Scale), o.X, o.Y)
	return c
}
