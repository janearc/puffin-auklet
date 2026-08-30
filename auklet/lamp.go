package auklet

import "time"

// Luxo Jr.: the third character, and the first with no face at all. The
// real lamp has never had eyes -- what reads as attention is the neck and
// shade turning, and what reads as personality is the hop. Both of those
// are Transform, which every emote in this package already has and none of
// them needed until now: the gopher and the puffin both read through a
// face, so nothing here forced Transform to carry a whole character by
// itself before.
//
// Built through NewSprite, same as the gopher. findEyes finds nothing in
// this art -- there are no E or X pixels anywhere in it -- so Gaze and
// WideEyes are both correctly inert, and buildEmotes' own guard against
// registering an all-empty performance quietly drops every stock emote that
// depended on an eye or a mouth (surprised, curious, shifty, sleepy,
// look-away, gasp). What survives for free is exactly what Transform alone
// can do: pop, settle, bob, nope. Blink does NOT survive derivation --
// there is no eye to shut -- so it is hand-authored below as the bulb
// going dark instead, the same "reuse what a blink means, not what it
// looks like" move the gopher's teeth already made with the beak roles.

// lampFlicker: the bulb (R, plus the Y halo immediately around it) painted
// dark. Coordinates read off auklet/lamp_front_art.go -- the bulb's
// bounding box is rows 4-11, cols 10-17.
var lampFlicker = Pose{{
	Name: "flicker", Part: PartOther, OX: 10, OY: 4,
	Art: []string{
		"KKKKKKKK", "KKKKKKKK", "KKKKKKKK", "KKKKKKKK",
		"KKKKKKKK", "KKKKKKKK", "KKKKKKKK", "KKKKKKKK",
	},
}}

// lampShadeAway: the shade's three bands, recentred toward the back corner
// of the poly instead of toward the viewer -- the bulb not facing you reads
// as "not looking at you yet," the same way a puffin not making eye contact
// is a gaze pose and not a different bird. Computed from the shade polygon
// in cmd/gen-art/lamp.go rather than eyeballed, the same way the astronaut
// helmets were.
var lampShadeAway = Pose{{
	Name: "shade-away", Part: PartOther, OX: 0, OY: 0,
	Art: []string{
		"............YB.......",
		"..........RRYYB......",
		".........RRRRYBB.....",
		".......RRRRRRYBBB....",
		"......YRRRRRRYBBBB...",
		"......YRRRRRRYBBBBB..",
		".......YRRRRYBBBBBBB.",
		".......YYYYYBBBBBBBB.",
		".......BBBBBBBBBBBB..",
		"........BBBBBBBBBBB..",
		"........BBBBBBBBBBB..",
		"........BBBBBBBB.....",
		".........BB..........",
		".....................",
	},
}}

// lampEmotes: hop is the character's whole signature, squat-spring-land in
// five frames of pure Transform. The four looks are one frame each, held
// briefly and not looping -- a glance, not a performance -- because at this
// size only the sign of the offset matters, the same rule Gaze uses for
// sprites that have a pupil to move instead of a body to lean.
func lampEmotes() []Emote {
	ms := func(n int) time.Duration { return time.Duration(n) * time.Millisecond }
	step := func(dx, dy int, scale float64) Transform { return Transform{DX: dx, DY: dy, Scale: scale} }
	look := func(name string, dx, dy int) Emote {
		return Emote{Name: name, Loop: false, Frames: []Frame{
			{Transform: step(dx, dy, 1), Hold: ms(380)},
			{Hold: ms(120)},
		}}
	}
	return []Emote{
		{
			Name: "hop", Loop: false,
			Frames: []Frame{
				{Transform: step(0, 1, 0.92), Hold: ms(90)},   // anticipation: squat
				{Transform: step(1, -3, 1.15), Hold: ms(110)}, // spring
				{Transform: step(1, -3, 1.08), Hold: ms(90)},  // airborne
				{Transform: step(1, 1, 0.95), Hold: ms(80)},   // land: compress
				{Transform: step(0, 0, 1), Hold: ms(120)},     // settle
			},
		},
		look("look-left", -2, 0),
		look("look-right", 2, 0),
		look("look-down", 0, 1),
		look("look-up", 0, -2),
		{
			// turn-in: leaning aside with the shade turned away, then the
			// turn itself -- shade-away drops off and the shade is facing
			// the fourth wall in the same beat the lean straightens out.
			// "Fourth wall" language stolen deliberately from the puffin's
			// own turn commit; the mechanism is smaller because there is
			// one view here, not four, but the beat is the same idea.
			Name: "turn-in", Loop: false,
			Frames: []Frame{
				{Pose: lampShadeAway, Transform: step(-3, 1, 0.95), Hold: ms(160)},
				{Pose: lampShadeAway, Transform: step(-2, 0, 0.98), Hold: ms(130)},
				{Transform: step(-1, -1, 1.05), Hold: ms(120)}, // the turn
				{Transform: step(0, 0, 1), Hold: ms(160)},
			},
		},
	}
}

func mustLampSprite(name string, art []string) Sprite {
	s, err := NewSprite(name, art, WithBlink(lampFlicker))
	if err != nil {
		panic("auklet: built-in lamp sprite " + name + ": " + err.Error())
	}
	for _, e := range lampEmotes() {
		WithEmote(e)(&s)
	}
	return s
}

// LampFrontView and LampFrontBallView: two Sprites, not one Sprite plus a
// pose, because the ball needs canvas room a pose cannot grow -- see
// cmd/gen-art/lamp.go. "Sometimes" the ball means picking the view, the
// same way y cycles turn angles for the other characters.
var LampFrontView = mustLampSprite("front", lampFrontArt)
var LampFrontBallView = mustLampSprite("front-ball", lampFrontBallArt)

func LampViews() []Sprite { return []Sprite{LampFrontView, LampFrontBallView} }
