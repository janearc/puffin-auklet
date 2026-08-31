package auklet

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
	Name string
	// Part is which region of the bird this overlay drives. It lets two things
	// animating at once be checked for collision rather than silently fighting
	// -- see Pose.Only.
	Part   Part
	OX, OY int      // top-left of the patch, in source pixels
	Art    []string // role codes; '.' means transparent, keep what is beneath

	// Replace marks this overlay as a full redraw of its own footprint, not
	// an addition to it: '.' cells inside it clear the destination to
	// RoleNone instead of leaving whatever the base had there. Everything
	// else still can only add, per the package doc -- this exists for
	// Ms. Pac-Man's spin, where each frame is a complete alternate drawing
	// and a gap where this frame has no bow must not let the base art's
	// own bow, still sitting at its resting position, show through.
	Replace bool
}

// Pose is a set of overlays applied in order. the zero Pose is the bird at rest.
type Pose []Overlay

// sideBlink closes the side view's eye: lid down, ring goes with it.
var sideBlink = Pose{{
	Name: "blink", Part: PartEyes, OX: 17, OY: 8,
	// the lid stops short of the nape stripe on purpose. run them together and
	// the two read as one long slot rather than as a closed eye.
	Art: []string{
		".WWWW.",
		".WWWW.",
		".DDDD.",
		".WWWW.",
	},
}}

// applyTo returns base with the pose stamped into it. it copies, so a sprite's
// art is never mutated -- a pose that leaked into the base would be a bird that
// never opened its eye again.
func (p Pose) applyTo(base []string) []string {
	if len(p) == 0 {
		return base
	}
	out := make([]string, len(base))
	buf := make([][]byte, len(base))
	for i, row := range base {
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
				if x < 0 || x >= len(buf[y]) {
					continue
				}
				if row[dx] == '.' {
					if o.Replace {
						buf[y][x] = '.'
					}
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

// frontBlink closes both eyes. two patches, not one: the eyes are far apart and
// a single overlay spanning them would repaint the beak between.
var frontBlink = Pose{
	{Name: "blink-l", Part: PartEyes, OX: 3, OY: 9, Art: []string{
		"WWWWW",
		"WWWWW",
		"DDDDD",
		"WWWWW",
	}},
	{Name: "blink-r", Part: PartEyes, OX: 16, OY: 9, Art: []string{
		"WWWWW",
		"WWWWW",
		"DDDDD",
		"WWWWW",
	}},
}
