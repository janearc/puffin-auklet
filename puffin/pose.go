package puffin

// Articulation, such as it is.
//
// the base art is one flat grid: every pixel belongs to exactly one part, and
// nothing is drawn behind anything else. that is enough to RECOLOUR a part --
// the role codes already segment the bird -- but not to MOVE one, because the
// pixels a part would uncover were never drawn.
//
// so the mechanism here is overlays, not bones. an Overlay is a small patch of
// role codes stamped over the base at a fixed offset, and a Pose is the set of
// overlays currently applied. that covers everything a mascot actually needs to
// do in a status bar -- blink, open its beak, look startled -- without anyone
// having to invent a rig.
//
// the stamping happens in ROLE space, before resampling and before any colour
// is chosen. that matters: a posed bird goes through exactly the same sampling
// and theming as a still one, so a pose cannot quietly break either.
type Overlay struct {
	Name   string
	OX, OY int      // top-left of the patch, in source pixels
	Art    []string // role codes; '.' means transparent, keep what is beneath
}

// Pose is a set of overlays applied in order. the zero Pose is the bird at rest.
type Pose []Overlay

// Blink closes the eye: lid down, and the orbital ring goes with it.
var Blink = Overlay{
	Name: "blink", OX: 17, OY: 8,
	// the lid stops short of the nape stripe on purpose. run them together and
	// the two read as one long slot rather than as a closed eye.
	Art: []string{
		".WWWW.",
		".WWWW.",
		".DDDD.",
		".WWWW.",
	},
}

// apply returns the base art with the pose stamped into it. it copies, so the
// package's art is never mutated -- a pose that leaked into the base would be a
// bird that never opened its eye again.
func (p Pose) apply() []string {
	if len(p) == 0 {
		return art
	}
	out := make([]string, len(art))
	buf := make([][]byte, len(art))
	for i, row := range art {
		buf[i] = []byte(row)
	}
	for _, o := range p {
		for dy, row := range o.Art {
			y := o.OY + dy
			if y < 0 || y >= len(buf) {
				continue
			}
			for dx := 0; dx < len(row); dx++ {
				x := o.OX + dx
				if x < 0 || x >= len(buf[y]) || row[dx] == '.' {
					continue
				}
				buf[y][x] = row[dx]
			}
		}
	}
	for i := range buf {
		out[i] = string(buf[i])
	}
	return out
}
