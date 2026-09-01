// The side view's stencil, the Theme its roles are painted with, and the
// validator that keeps a theme legible. See doc.go for the package overview.
//
// The art is a fixed 39x44 pixel grid of role codes. Colours are supplied by
// the caller; the art itself never names one, so retheming is a struct rather
// than a redraw.

package auklet

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/janearc/puffin-auklet/canvas"
)

// Width and Height are the side view's footprint at native size, in cells.
const (
	Width  = 39
	Height = 22
)

// Theme assigns a colour to each part of the bird. roles are named for what
// they are, not what colour they usually are, so a theme can recolour the
// puffin without the art knowing.
//
// the roles are not independent. Validate reports the pairs that must stay
// distinguishable; a theme that fails it renders a blob.
type Theme struct {
	// Background fills the pixels outside the bird. leave it nil to make the
	// sprite a CUTOUT -- those cells are then skipped by canvas.Blit and
	// whatever is behind shows through, including through the half-cells along
	// the silhouette's edge.
	Background lipgloss.TerminalColor

	Dark  lipgloss.TerminalColor // crown, back, throat collar
	Light lipgloss.TerminalColor // face patch and belly
	Wing  lipgloss.TerminalColor // tone on the folded wing, sits just off Dark

	BeakBase lipgloss.TerminalColor // slate section against the face
	BeakBand lipgloss.TerminalColor // the ridge stripe and the gape line
	BeakTip  lipgloss.TerminalColor // the outer half

	Feet    lipgloss.TerminalColor
	Pupil   lipgloss.TerminalColor
	EyeRing lipgloss.TerminalColor // orbital ring
	Stripe  lipgloss.TerminalColor // eye plate and the line trailing to the nape
}

// DefaultTheme is the bird as drawn: an atlantic puffin on a near-black field.
func DefaultTheme() Theme {
	return Theme{
		Background: lipgloss.Color("#1e222b"),
		Dark:       lipgloss.Color("#141418"),
		Light:      lipgloss.Color("#f4f6f8"),
		Wing:       lipgloss.Color("#2b2e36"),
		BeakBase:   lipgloss.Color("#6b748c"),
		BeakBand:   lipgloss.Color("#f0b42a"),
		BeakTip:    lipgloss.Color("#e03c1e"),
		Feet:       lipgloss.Color("#ff7a1a"),
		Pupil:      lipgloss.Color("#0a0a0c"),
		EyeRing:    lipgloss.Color("#c82418"),
		Stripe:     lipgloss.Color("#3a3d45"),
	}
}

// WithBackground returns a copy of t sitting on bg. use it to drop the bird
// into a panel whose background differs from the app's.
func (t Theme) WithBackground(bg lipgloss.TerminalColor) Theme {
	t.Background = bg
	return t
}

func (t Theme) colorFor(c byte) lipgloss.TerminalColor {
	switch c {
	case 'K':
		return t.Dark
	case 'W':
		return t.Light
	case 'V':
		return t.Wing
	case 'B':
		return t.BeakBase
	case 'Y':
		return t.BeakBand
	case 'R':
		return t.BeakTip
	case 'O':
		return t.Feet
	case 'E':
		return t.Pupil
	case 'X':
		return t.EyeRing
	case 'D':
		return t.Stripe
	case '_':
		// A HOLE. Not a colour -- the field, whatever the field is.
		//
		// Ms Pac-Man's open mouth was 'E', the pupil, and the theme said so:
		// "the eye, and the open mouth's interior". On a cabinet that is
		// right, because the cabinet is black and a black mouth and an
		// absent one are the same picture. Off a cabinet they stop being
		// the same picture: puffin nils Background for the cutout, so a
		// mouth drawn in pupil-black is a black wedge stuck to a yellow
		// disc on a purple page.
		//
		// '_' resolves to Background like '.' does, so it is cabinet-black
		// in bandersnatch and transparent in a cutout -- the arcade reading
		// and the correct one, without choosing between them. It is NOT
		// '.', because '.' means "leave what is beneath" when an overlay
		// composites, and the thing beneath the mouth is the disc.
		return t.Background
	default:
		return t.Background
	}
}

// SizeAt reports the widget's footprint in cells at a given integer scale.
func SizeAt(scale int) (w, h int) {
	if scale < 1 {
		scale = 1
	}
	return Width * scale, Height * scale
}

// Cells returns the sprite at an integer scale using half blocks. it is a
// thin wrapper on CellsAt, kept because scale 1 is the art's native size and
// most callers want exactly that.
func Cells(t Theme, scale int) [][]canvas.Cell {
	if scale < 1 {
		scale = 1
	}
	return CellsAt(t, Half, Width*scale, Height*scale)
}

// Render returns the puffin as a printable string at scale 1, Height lines
// tall with no trailing newline. it does not validate; call Validate once at
// theme load.
func Render(t Theme) string { return canvas.Render(Cells(t, 1)) }

// --- theme validation ---------------------------------------------------
//
// the puffin is legible because of contrast relationships, not hues. recolour
// it freely, but these pairs have to stay apart or the silhouette collapses.

type rgb struct{ r, g, b float64 }

// resolve pulls declared values out of a TerminalColor. it deliberately does
// not use RGBA(): that answer is filtered through the renderer's colour
// profile and returns black when there is no tty, which would make every
// theme look identical and pass.
//
// an adaptive colour is checked on both faces. a palette-index colour
// ("205", "9") carries no rgb we can read here, so it is reported as
// unverifiable rather than silently passed.
func resolve(c lipgloss.TerminalColor) ([]rgb, bool) {
	var out []rgb
	add := func(s string) bool {
		v, ok := parseHex(s)
		if ok {
			out = append(out, v)
		}
		return ok
	}
	switch v := c.(type) {
	case lipgloss.Color:
		if !add(string(v)) {
			return nil, false
		}
	case lipgloss.AdaptiveColor:
		if !add(v.Light) || !add(v.Dark) {
			return nil, false
		}
	case lipgloss.CompleteColor:
		if !add(v.TrueColor) {
			return nil, false
		}
	case lipgloss.CompleteAdaptiveColor:
		if !add(v.Light.TrueColor) || !add(v.Dark.TrueColor) {
			return nil, false
		}
	default:
		return nil, false
	}
	return out, true
}

func parseHex(s string) (rgb, bool) {
	s = strings.TrimPrefix(s, "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return rgb{}, false
	}
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return rgb{}, false
	}
	return rgb{
		float64((n>>16)&0xff) / 255,
		float64((n>>8)&0xff) / 255,
		float64(n&0xff) / 255,
	}, true
}

// dist is plain euclidean distance in srgb, 0..~1.73. good enough to answer
// "can a person tell these two cells apart" without dragging in a colour
// science dependency.
func dist(a, b rgb) float64 {
	dr, dg, db := a.r-b.r, a.g-b.g, a.b-b.b
	return sqrt(dr*dr + dg*dg + db*db)
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 24; i++ {
		z = 0.5 * (z + x/z)
	}
	return z
}

func lum(c rgb) float64 {
	lin := func(v float64) float64 {
		if v <= 0.03928 {
			return v / 12.92
		}
		p := (v + 0.055) / 1.055
		return p * p * p // close enough to ^2.4 for a legibility gate
	}
	return 0.2126*lin(c.r) + 0.7152*lin(c.g) + 0.0722*lin(c.b)
}

type pair struct {
	aName, bName string
	a, b         lipgloss.TerminalColor
	min          float64
	why          string
}

// Validate reports every role pair that has collapsed, all at once. a theme
// that fails this still renders; it just renders something that is not
// identifiably a auklet.
func (t Theme) Validate() error {
	var errs []error

	// the tuxedo contrast is the bird. this one is a luminance gate, not a
	// distance gate: a mid-grey and a mid-blue are far apart in rgb and still
	// read as one flat mass at cell size.
	if lo, ok1 := resolve(t.Dark); ok1 {
		if hi, ok2 := resolve(t.Light); ok2 {
			for _, a := range lo {
				for _, b := range hi {
					if lum(b)-lum(a) < 0.30 {
						errs = append(errs, fmt.Errorf(
							"Light is not bright enough against Dark (luminance gap %.3f, need 0.300): the face and belly stop separating from the cap",
							lum(b)-lum(a)))
					}
				}
			}
		}
	}

	checks := []pair{
		{"Dark", "Background", t.Dark, t.Background, 0.08,
			"the silhouette disappears into the field"},
		{"BeakTip", "BeakBand", t.BeakTip, t.BeakBand, 0.15,
			"the beak stops reading as banded, which is the puffin's whole signature"},
		{"BeakBand", "BeakBase", t.BeakBand, t.BeakBase, 0.15,
			"the beak stops reading as banded"},
		{"BeakTip", "BeakBase", t.BeakTip, t.BeakBase, 0.15,
			"the beak flattens to one colour"},
		{"BeakBase", "Light", t.BeakBase, t.Light, 0.15,
			"the beak merges into the face and the head loses its front edge"},
		{"BeakTip", "Background", t.BeakTip, t.Background, 0.15,
			"the beak tip is cut off by the field"},
		{"Feet", "Background", t.Feet, t.Background, 0.15,
			"the bird floats"},
		{"Pupil", "EyeRing", t.Pupil, t.EyeRing, 0.15,
			"the eye becomes a solid dot"},
		{"EyeRing", "Light", t.EyeRing, t.Light, 0.15,
			"the eye vanishes into the face"},
		{"Wing", "Dark", t.Wing, t.Dark, 0.02,
			"the wing tone is doing nothing; drop it or widen it"},
	}

	for _, p := range checks {
		if t.Background == nil && (p.aName == "Background" || p.bName == "Background") {
			// a cutout has no field of its own. whether the bird reads against
			// what is behind it is the compositor's problem, and nothing here
			// can see that surface.
			continue
		}
		as, ok1 := resolve(p.a)
		bs, ok2 := resolve(p.b)
		if !ok1 || !ok2 {
			errs = append(errs, fmt.Errorf(
				"%s/%s: cannot verify (a palette-index or custom colour carries no readable rgb); check this pair by eye",
				p.aName, p.bName))
			continue
		}
		worst := 99.0
		for _, a := range as {
			for _, b := range bs {
				if d := dist(a, b); d < worst {
					worst = d
				}
			}
		}
		if worst < p.min {
			errs = append(errs, fmt.Errorf("%s and %s are too close (%.3f, need %.3f): %s",
				p.aName, p.bName, worst, p.min, p.why))
		}
	}

	return errors.Join(errs...)
}

// Luminance reports the relative luminance of c, 0..1, and whether it could be
// read at all. adapters need it: a design system names its colours for their
// role in a page ("ink" is text), not for how bright they are, so mapping one
// onto the puffin's light/dark roles means asking.
//
// an adaptive colour reports its dark face.
func Luminance(c lipgloss.TerminalColor) (float64, bool) {
	vs, ok := resolve(c)
	if !ok || len(vs) == 0 {
		return 0, false
	}
	return lum(vs[len(vs)-1]), true
}

// RoleColor maps one role byte to its colour under this theme. '.' and any
// unknown byte return Background, which may be nil for a cutout.
func (t Theme) RoleColor(role byte) lipgloss.TerminalColor { return t.colorFor(role) }

// RGB resolves a colour to eight-bit components. it reads the declared value
// rather than asking the renderer, for the same reason Validate does: the
// renderer's answer depends on a terminal that an image encoder does not have.
func RGB(c lipgloss.TerminalColor) (r, g, b uint8, ok bool) {
	vs, ok := resolve(c)
	if !ok || len(vs) == 0 {
		return 0, 0, 0, false
	}
	v := vs[len(vs)-1]
	return uint8(v.r*255 + 0.5), uint8(v.g*255 + 0.5), uint8(v.b*255 + 0.5), true
}
