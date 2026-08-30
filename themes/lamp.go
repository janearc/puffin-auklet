package themes

import "github.com/janearc/puffin-auklet/auklet"

// The lamp's own colours: brushed graphite for the body, a warm desk-cream
// field rather than the gopher's dark navy or the puffin's cutout default --
// a lamp's native surface is a desk, so it gets one. Reference theme, the
// same way atlantic and the gopher theme are reference rather than dodo.
func init() {
	All = append(All, Named{
		Name: "lamp", Note: "the lamp's own colours, on a desk",
		Theme: auklet.Theme{
			Background: c("#E8DCC8"),
			Dark:       c("#5B6470"), // brushed graphite: arm, base, joints
			Light:      c("#F4F2EE"), // unused by the art itself; still has to validate
			Wing:       c("#9AA3AD"), // the highlight arc on the base
			BeakBase:   c("#3E434B"), // shade rim, darker than the body
			BeakBand:   c("#9C8B6E"), // shade mid-band, warm brass-gray
			BeakTip:    c("#FFB627"), // the bulb, glowing
			Feet:       c("#6B7280"), // unused -- see cmd/gen-art/lamp.go
			Pupil:      c("#14171B"),
			EyeRing:    c("#8B93A0"),
			Stripe:     c("#23262B"), // the joints' rivet dots
		},
	})
}
