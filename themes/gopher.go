package themes

import "github.com/janearc/puffin-auklet/auklet"

// The gopher's own colours, the way "atlantic" is the puffin's own colours:
// a reference theme rather than one of dodo's four, because the gopher did
// not come from dodo's stylesheet and has no page to borrow ink from.
//
// Appended in init rather than added to All directly, so this file is the
// only thing that has to exist for the gopher to get a native theme --
// themes.go does not need to know this file exists.
func init() {
	All = append(All, Named{
		Name: "gopher", Note: "the gopher's own colours, as drawn",
		Theme: auklet.Theme{
			Background: c("#12161C"),
			Dark:       c("#00ADD8"), // Go's own blue, not an invented teal
			Light:      c("#FFFFFF"), // eye whites
			Wing:       c("#00829E"), // a shade off it, down each flank
			BeakBase:   c("#F2C88F"), // the tan muzzle
			BeakBand:   c("#6B4226"), // the gumline
			BeakTip:    c("#FFFDF5"), // the teeth
			Feet:       c("#F2C88F"), // paws and feet, same tan as the muzzle
			Pupil:      c("#101317"),
			EyeRing:    c("#B9DEE3"), // a hair off white, so the ring still reads as a ring
			Stripe:     c("#101317"), // the nose, and the inner-ear mark
		},
	})
}
