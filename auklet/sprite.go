package auklet

import "math"

// A Sprite is one view of the bird: a stencil of role codes plus the poses that
// belong to that view.
//
// poses are per-sprite because they are coordinates, not concepts. "blink" in
// profile is one eyelid at one place; head-on it is two, somewhere else. an
// overlay from the wrong view lands on the wrong pixels and the bird gets a
// lid across its cheek.
type Sprite struct {
	Name  string
	art   []string
	Blink Pose

	// Mouths are the mouth shapes, closed first, widening. Mouths[0] is the
	// beak shut and is empty by construction: the base art is already closed.
	Mouths []Pose
}

// Size reports the stencil's dimensions in source pixels.
func (s Sprite) Size() (w, h int) { return len(s.art[0]), len(s.art) }

// ColsFor returns the column count that preserves this sprite's proportions at
// a given row count.
func (s Sprite) ColsFor(rows int) int {
	w, h := s.Size()
	return int(math.Round(float64(rows) * 2 * float64(w) / float64(h)))
}

// Mouth returns the pose for an openness level, clamped. level 0 is shut.
func (s Sprite) Mouth(level int) Pose {
	if len(s.Mouths) == 0 {
		return nil
	}
	if level < 0 {
		level = 0
	}
	if level >= len(s.Mouths) {
		level = len(s.Mouths) - 1
	}
	return s.Mouths[level]
}

// MouthLevels is how many shapes this sprite's mouth has, counting shut.
func (s Sprite) MouthLevels() int {
	if len(s.Mouths) == 0 {
		return 1
	}
	return len(s.Mouths)
}

// Views is every sprite, in the order a viewer should cycle them.
func Views() []Sprite { return []Sprite{SideView, FrontView} }

// SideView is the bird in profile: the classic silhouette, and the one that
// reads at the smallest sizes because the beak is doing all the work.
//
// it has no mouth shapes. opening a beak side-on means swinging the lower
// mandible away from the head, and the head behind it was never drawn.
var SideView = Sprite{Name: "side", art: sideArt, Blink: sideBlink}

// FrontView is the bird facing the reader. head-on the beak is a narrow blade
// rather than the profile's wedge -- it is flattened side to side, so the
// feature that makes a puffin a puffin is the thing that mostly disappears.
// what carries the read instead is the symmetry: two white cheeks split by that
// blade, two ringed eyes set high and close.
//
// this is the view that can talk.
var FrontView = Sprite{
	Name: "front", art: frontArt, Blink: frontBlink,
	Mouths: []Pose{
		nil, // shut
		{{Name: "ajar", OX: 10, OY: 17, Art: []string{"KKKK"}}},
		{{Name: "open", OX: 10, OY: 17, Art: []string{"KKKK", "KKKK"}}},
		{
			{Name: "wide", OX: 10, OY: 16, Art: []string{
				"KKKK", "KKKK", "KKKK", "KKKK",
			}},
			// the lower mandible swings down, so the beak gets longer
			{Name: "wide-jaw", OX: 11, OY: 24, Art: []string{"RR", "RR"}},
		},
	},
}

// Pixels returns this sprite's role grid with a pose stamped in: one byte per
// source pixel, '.' meaning transparent.
//
// it is the escape hatch for renderers that are not terminals. the whole design
// leans on colour being chosen last, so anything that can map a role byte to a
// colour -- an image encoder, a web canvas, a plotter -- can draw this bird
// without the puffin package knowing that target exists.
func (s Sprite) Pixels(pose Pose) []string { return pose.applyTo(s.art) }
