package auklet

import "time"

// Ms. Pac-Man: a circle, a bow, and a mouth that is the whole character.
//
// The wedge mouth is an overlay stack, not a hole cut in the base art, for
// the same reason the gopher's teeth and the puffin's beak both are: an
// overlay can only ADD pixels, an already-drawn disc cannot be un-drawn
// from outside it, and a pose that could leak into the base would be a
// mouth that never closed again. The base art (mustMsPacmanSprite's own
// art) is drawn shut -- a full, uncut disc -- and each mouth level paints a
// progressively wider pie-slice of RolePupil (E) over the right side of it.
// E, not a fresh role: a real Pac-Man's eye dot and open-mouth interior are
// conventionally the same near-black, and the alphabet does not owe this
// subject a new slot for a colour it already has one for.
//
// Coordinates read off the actual generated art (cmd/calcpac, throwaway,
// deleted) rather than eyeballed -- the same way the astronaut helmets and
// the lamp's shade-away were.

// level 1: half-angle 14 degrees
var msPacmanMouth1 = Pose{{
	Name: "mouth", Part: PartMouth, OX: 5, OY: 6,
	Art: []string{
		"......................",
		"......................",
		"......................",
		"......................",
		"......................",
		"......................",
		"......................",
		"......................",
		"......................",
		"......................",
		"..................EEE.",
		"..............EEEEEEE.",
		"..........EEEEEEEEEEE.",
		"..........EEEEEEEEEEE.",
		"..............EEEEEEE.",
		"..................EEE.",
		"......................",
		"......................",
		"......................",
		"......................",
		"......................",
		"......................",
		"......................",
		"......................",
		"......................",
		"......................",
	},
}}

// level 2: half-angle 28 degrees
var msPacmanMouth2 = Pose{{
	Name: "mouth", Part: PartMouth, OX: 5, OY: 6,
	Art: []string{
		"......................",
		"......................",
		"......................",
		"......................",
		"......................",
		"......................",
		"......................",
		"..................EE..",
		"................EEEE..",
		"...............EEEEEE.",
		".............EEEEEEEE.",
		"...........EEEEEEEEEE.",
		".........EEEEEEEEEEEE.",
		".........EEEEEEEEEEEE.",
		"...........EEEEEEEEEE.",
		".............EEEEEEEE.",
		"...............EEEEEE.",
		"................EEEE..",
		"..................EE..",
		"......................",
		"......................",
		"......................",
		"......................",
		"......................",
		"......................",
		"......................",
	},
}}

// level 3: half-angle 44 degrees -- as wide as the arcade original opens.
var msPacmanMouth3 = Pose{{
	Name: "mouth", Part: PartMouth, OX: 5, OY: 6,
	Art: []string{
		"......................",
		"......................",
		"......................",
		"......................",
		".................E....",
		"................EEE...",
		"...............EEEE...",
		"..............EEEEEE..",
		".............EEEEEEE..",
		"............EEEEEEEEE.",
		"...........EEEEEEEEEE.",
		"..........EEEEEEEEEEE.",
		".........EEEEEEEEEEEE.",
		".........EEEEEEEEEEEE.",
		"..........EEEEEEEEEEE.",
		"...........EEEEEEEEEE.",
		"............EEEEEEEEE.",
		".............EEEEEEE..",
		"..............EEEEEE..",
		"...............EEEE...",
		"................EEE...",
		".................E....",
		"......................",
		"......................",
		"......................",
		"......................",
	},
}}

// msPacmanEmotes: chomp is the character, looping, so it can idle on it
// without being triggered -- Jane asked for exactly this. waka is the same
// bite cycle with a steady rightward drift layered on, the arcade's own
// "moving while eating" read, and the reason this character was worth
// building: a circle is the one shape here where a whole-sprite Transform
// reads as motion rather than as a wobble, because rotating it changes
// nothing a viewer would notice missing.
func msPacmanEmotes() []Emote {
	ms := func(n int) time.Duration { return time.Duration(n) * time.Millisecond }
	bite := func(dx int) []Frame {
		d := Transform{DX: dx}
		return []Frame{
			{Transform: d, Hold: ms(120)},
			{Pose: msPacmanMouth2, Transform: d, Hold: ms(90)},
			{Pose: msPacmanMouth3, Transform: d, Hold: ms(110)},
			{Pose: msPacmanMouth2, Transform: d, Hold: ms(90)},
		}
	}
	return []Emote{
		{Name: "chomp", Loop: true, Frames: bite(0)},
		{Name: "waka", Loop: true, Frames: bite(1)},
	}
}

func mustMsPacmanSprite(name string, art []string) Sprite {
	s, err := NewSprite(name, art, WithMouths(
		nil, // shut: the base art's disc, uncut
		msPacmanMouth1,
		msPacmanMouth2,
		msPacmanMouth3,
	))
	if err != nil {
		panic("auklet: built-in mspacman sprite " + name + ": " + err.Error())
	}
	for _, e := range msPacmanEmotes() {
		WithEmote(e)(&s)
	}
	return s
}

var MsPacmanFrontView = mustMsPacmanSprite("front", msPacmanFrontArt)

func MsPacmanViews() []Sprite { return []Sprite{MsPacmanFrontView} }
