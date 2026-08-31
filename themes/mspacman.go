package themes

import "github.com/janearc/puffin-auklet/auklet"

// Ms. Pac-Man's own colours: arcade yellow on a cabinet-black field, the
// bow in reds and pinks. Reference theme, the same way atlantic, the
// gopher theme and the lamp theme all are.
func init() {
	All = append(All, Named{
		Name: "mspacman", Note: "the arcade's own colours, on cabinet black",
		Theme: auklet.Theme{
			Background: c("#0B0B1E"),
			Dark:       c("#D9AE00"), // the disc -- deep enough that Light still clears the luminance gate
			Light:      c("#EAF4FF"), // unused by the art itself; still has to validate
			Wing:       c("#FFE066"), // the highlight crescent, a shade lighter than the disc
			BeakBase:   c("#B0164F"), // bow rim
			BeakBand:   c("#F06292"), // bow fill
			BeakTip:    c("#E53935"), // the knot -- the bow's accent
			Feet:       c("#EF5350"), // the shoes
			Pupil:      c("#0D0D0D"), // the eye, and the open mouth's interior
			EyeRing:    c("#FFF3C4"),
			Stripe:     c("#1A1A1A"), // the beauty mark
		},
	})
}
