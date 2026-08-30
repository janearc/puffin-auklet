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

// LampFrontView is the only view so far -- the arm's zigzag and the shade's
// tilt are both drawn facing the camera already, and a profile of a lamp is
// a much smaller design question than a profile of a face. Extend later if
// it turns out to want one.
var LampFrontView = mustLampSprite("front", lampFrontArt)

func LampViews() []Sprite { return []Sprite{LampFrontView} }
