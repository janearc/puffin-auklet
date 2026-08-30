package themes

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/janearc/puffin-auklet/auklet"
)

// Vars is the subset of a dodo theme the puffin adapter reads. mirrored from
// dodo/lib/themes/themes.css by cmd/gen-themes; see dodo_gen.go.
type Vars struct {
	Bg, Panel, Raised, Line, Ink, Dim, Accent, AccentToken string
}

// Puffin maps a dodo theme onto the bird.
//
// the mapping is not a rename. dodo names its colours for their job in a PAGE
// -- "ink" is body text, "bg" is the ground -- so on a dark theme ink is nearly
// white and on stardew it is nearly black. the puffin needs LUMINANCE roles.
// so the adapter branches on the ground's brightness and picks accordingly;
// a naive ink->Light mapping renders stardew as a dark bird with a darker face.
//
// semantic colours (--ok, --warn, --caution) are deliberately not used. a
// warning colour spent on decoration stops reading as a warning.
func (v Vars) Auklet() auklet.Theme {
	c := func(s string) lipgloss.Color { return lipgloss.Color(s) }

	lum, _ := auklet.Luminance(c(v.Bg))
	lightGround := lum > 0.18

	dark, light, wing := v.Raised, v.Ink, v.Line
	if lightGround {
		dark, light, wing = v.Ink, v.Panel, v.Dim
	}

	return auklet.Theme{
		Background: c(v.Bg),
		Dark:       c(dark),
		Light:      c(light),
		Wing:       c(wing),

		// the tip is the largest area of the beak, so it carries the theme's
		// signature colour; the ridge takes the complementary token.
		BeakBase: c(v.Dim),
		BeakBand: c(v.AccentToken),
		BeakTip:  c(v.Accent),

		Feet:    c(v.Accent),
		Pupil:   c(dark),
		EyeRing: c(v.Accent),
		Stripe:  c(v.Dim),
	}
}
