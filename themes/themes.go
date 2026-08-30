// package themes is the theme set for the puffin viewer.
//
// dodo's four come first and come from dodo's own themes.css via
// cmd/gen-themes -- the hex is not duplicated here. after them sit a reference
// theme and three diagnostics, which exist to exercise the validator: a picker
// that only ever shows passing themes teaches you nothing about the gate.
package themes

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/janearc/puffin-auklet/auklet"
)

type Named struct {
	Name  string
	Note  string
	Theme auklet.Theme
}

func c(s string) lipgloss.Color { return lipgloss.Color(s) }

func dodo(name, note string) Named {
	return Named{Name: name, Note: note, Theme: Dodo[name].Auklet()}
}

var All = []Named{
	dodo("corvid", "dodo default"),
	dodo("vaporwave", "dodo"),
	dodo("bladerunner", "dodo"),
	{
		Name: "atlantic", Note: "the reference bird, as drawn",
		Theme: auklet.DefaultTheme(),
	},
	{
		Name: "nord", Note: "plausible, and it fails -- cap sinks into the field",
		Theme: auklet.Theme{
			Background: c("#2e3440"), Dark: c("#242933"), Light: c("#eceff4"),
			Wing: c("#3b4252"), BeakBase: c("#5e81ac"), BeakBand: c("#ebcb8b"),
			BeakTip: c("#bf616a"), Feet: c("#d08770"), Pupil: c("#2e3440"),
			EyeRing: c("#bf616a"), Stripe: c("#4c566a"),
		},
	},
	{
		Name: "ansi16", Note: "palette indices -- unverifiable by design",
		Theme: auklet.Theme{
			Background: c("8"), Dark: c("0"), Light: c("15"), Wing: c("0"),
			BeakBase: c("4"), BeakBand: c("11"), BeakTip: c("9"),
			Feet: c("9"), Pupil: c("0"), EyeRing: c("1"), Stripe: c("8"),
		},
	},
	{
		Name: "collapsed", Note: "deliberately broken",
		Theme: auklet.Theme{
			Background: c("#141418"), Dark: c("#141418"), Light: c("#8a8f99"),
			Wing: c("#141418"), BeakBase: c("#f0b42a"), BeakBand: c("#f0b42a"),
			BeakTip: c("#f0b42a"), Feet: c("#1b1e24"), Pupil: c("#c82418"),
			EyeRing: c("#c82418"), Stripe: c("#141418"),
		},
	},
	dodo("stardew", "dodo's light theme -- the luminance branch, kept honest"),
}
