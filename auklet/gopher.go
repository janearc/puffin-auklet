package auklet

import (
	"strings"
	"time"
)

// The Go gopher: the second character in this package, built entirely
// through NewSprite -- the same "way in for other people's birds" path a
// stranger's sprite file goes through. Blink, gaze, wide-eyes and the whole
// stock emote vocabulary come from that path for free; only the mouth is
// hand-authored here, because Mouths is the one thing NewSprite has no
// opinion about.
//
// Nothing in this file touches sprite.go, pose.go or any puffin art. A
// second character living here is the point of NewSprite existing at all.

// gopherMouthPose builds one mouth level: the gumline widens into a dark gap
// of gapRows, and the teeth drop to the bottom of it. Coordinates are read
// directly off auklet/gopher_front_art.go -- the gumline sits at row 28,
// cols 11-19, and the two teeth at cols 13-14 and 16-17.
func gopherMouthPose(gapRows int) Pose {
	const (
		ox, oy = 11, 28
		w      = 9 // cols 11..19
		toothL = 2 // col 13, relative to ox
		toothR = 5 // col 16, relative to ox
	)
	var rows []string
	for i := 0; i < gapRows; i++ {
		rows = append(rows, strings.Repeat(string(RoleDark), w))
	}
	row := []byte(strings.Repeat(string(RoleDark), w))
	row[toothL], row[toothL+1] = RoleBeakTip, RoleBeakTip
	row[toothR], row[toothR+1] = RoleBeakTip, RoleBeakTip
	for i := 0; i < 4; i++ {
		rows = append(rows, string(row))
	}
	return Pose{{Name: "mouth", Part: PartMouth, OX: ox, OY: oy, Art: rows}}
}

// The wave: the right arm's stub (rows 36-39, cols 26-30 in
// auklet/gopher_front_art.go) is EXTENDED upward rather than relocated. An
// overlay can only add pixels, never clear a base one back to transparent --
// so the trick that works here is the same one WideEyes uses on the puffin:
// grow a part from where it already is, into space that was already empty,
// instead of trying to erase and repaint it somewhere else. Cols 26-30 are
// clear all the way up to the eyes (col 25 is the last face pixel at every
// row from 18 to 39), so the raised arm never overlaps the face except by
// one column, for six rows, in the tilted frame.
var gopherWaveUp = Pose{{
	Name: "wave-up", Part: PartOther, OX: 26, OY: 18,
	Art: []string{
		"OOOOO", "OOOOO", "OOOOO", "OOOOO", "OOOOO", "OOOOO",
		"OOOOO", "OOOOO", "OOOOO", "OOOOO", "OOOOO", "OOOOO",
		"OOOOO", "OOOOO", "OOOOO", "OOOOO", "OOOOO", "OOOOO",
	},
}}

var gopherWaveTilt = Pose{{
	Name: "wave-tilt", Part: PartOther, OX: 25, OY: 18,
	Art: []string{
		// the hand: swung one column in, toward the head
		"OOOOO", "OOOOO", "OOOOO", "OOOOO", "OOOOO", "OOOOO",
		// the forearm: planted, same column as the base stub
		".OOOOO", ".OOOOO", ".OOOOO", ".OOOOO", ".OOOOO", ".OOOOO",
		".OOOOO", ".OOOOO", ".OOOOO", ".OOOOO", ".OOOOO", ".OOOOO",
	},
}}

// gopherWalkIn is an entrance: the sprite grows and bobs as if walking up to
// the camera -- scale is how "closer" reads on a sprite that is already
// facing the fourth wall, the same way it already does for "pop" -- then
// plants and waves before settling back to rest. One shot, not a loop: an
// entrance that repeated forever would stop being one.
func gopherWalkIn() Emote {
	ms := func(n int) time.Duration { return time.Duration(n) * time.Millisecond }
	step := func(dx, dy int, scale float64) Transform { return Transform{DX: dx, DY: dy, Scale: scale} }
	return Emote{
		Name: "walk-in",
		Frames: []Frame{
			{Transform: step(-1, 1, 0.85), Hold: ms(120)},
			{Transform: step(1, 0, 0.90), Hold: ms(120)},
			{Transform: step(-1, 1, 0.95), Hold: ms(120)},
			{Transform: step(1, 0, 1.00), Hold: ms(120)},
			{Transform: step(-1, 1, 1.06), Hold: ms(120)},
			{Transform: step(0, 0, 1.12), Hold: ms(140)},
			{Pose: gopherWaveUp, Transform: step(0, 0, 1.12), Hold: ms(160)},
			{Pose: gopherWaveTilt, Transform: step(0, 0, 1.12), Hold: ms(160)},
			{Pose: gopherWaveUp, Transform: step(0, 0, 1.12), Hold: ms(160)},
			{Pose: gopherWaveTilt, Transform: step(0, 0, 1.12), Hold: ms(160)},
			{Transform: step(0, 0, 1.05), Hold: ms(200)},
		},
	}
}

// Pointing: three more lengths of the same right-arm extension the wave
// uses, plus a mirrored one on the left arm. None of this is new art in the
// sense of new pixels invented from nothing -- it is the same stub, grown a
// different amount in a different direction, which is the whole trick this
// package has for moving anything.
var gopherPointDown = Pose{{
	Name: "point-down", Part: PartOther, OX: 26, OY: 36,
	Art: []string{"OOOOO", "OOOOO", "OOOOO", "OOOOO", "OOOOO", "OOOOO", "OOOOO"},
}}

var gopherPointRight = Pose{{
	Name: "point-right", Part: PartOther, OX: 26, OY: 28,
	Art: []string{
		"OOOOO", "OOOOO", "OOOOO", "OOOOO",
		"OOOOO", "OOOOO", "OOOOO", "OOOOO",
		"OOOOO", "OOOOO", "OOOOO", "OOOOO",
	},
}}

// the left arm's stub (cols 0-4, rows 36-39) is transparent above it too,
// the same as the right's -- confirmed against gopher_front_art.go rather
// than assumed from the mirrored draw code.
var gopherPointLeft = Pose{{
	Name: "point-left", Part: PartOther, OX: 0, OY: 28,
	Art: []string{
		"OOOOO", "OOOOO", "OOOOO", "OOOOO",
		"OOOOO", "OOOOO", "OOOOO", "OOOOO",
		"OOOOO", "OOOOO", "OOOOO", "OOOOO",
	},
}}

// Brows: the gopher's base art has none, so these are additive above the
// eyes rather than a repaint of something already there. Two overlays per
// pose, same reason the puffin's blink is two overlays -- the eyes are far
// enough apart that one patch spanning both would repaint the nose bridge
// between them.
// A one-row tilt turned out not to read at this resolution -- angry and
// sad rendered indistinguishably. Position and weight carry further than
// angle does at eight-odd rows: angry presses a thick brow down onto the
// eye; sad lifts a thin one well clear of it. Four rows of vertical
// separation is the whole trick.
var gopherAngryBrow = Pose{
	{Name: "brow-l", Part: PartEyes, OX: 7, OY: 8, Art: []string{"DDDDD", "DDDDD"}},
	{Name: "brow-r", Part: PartEyes, OX: 19, OY: 8, Art: []string{"DDDDD", "DDDDD"}},
}

var gopherSadBrow = Pose{
	{Name: "brow-l", Part: PartEyes, OX: 7, OY: 5, Art: []string{"DDDDD"}},
	{Name: "brow-r", Part: PartEyes, OX: 19, OY: 5, Art: []string{"DDDDD"}},
}

// gopherHelmet: a full ring one to two pixels proud of the head silhouette,
// computed from the same ellipse the head capsule's top cap uses (center
// 15,11, radius 10 -- see capsule() in cmd/gen-art/gopher.go) rather than
// eyeballed. Role X, reused here as glass/rim tone the way every role gets
// reused across subjects.
var gopherHelmet = Pose{{
	Name: "astronaut", Part: PartOther, OX: 2, OY: -1,
	Art: []string{
		"..........XXXXXX...........",
		".......XXXXXXXXXXXX........",
		"......XXX........XXX.......",
		".....XX............XX......",
		"....XX..............XX.....",
		"...XX................XX....",
		"..XX..................XX...",
		"..XX..................XX...",
		"..X....................X...",
		".XX....................XX..",
		".XX....................XX..",
		".XX....................XX..",
		".XX....................XX..",
		".XX....................XX..",
		".XX....................XX..",
		"..X....................X...",
		"..XX..................XX...",
		"..XX..................XX...",
		"...XX................XX....",
		"....XX..............XX.....",
		".....XX............XX......",
		"......XXX........XXX.......",
		".......XXXXXXXXXXXX........",
		"..........XXXXXX...........",
		"...........................",
	},
}}

// gopherLaughing needs the sprite's own (derived) blink, so it is built
// after NewSprite returns rather than passed in as an option -- Blink does
// not exist yet while the option list that constructs the sprite is still
// running.
func gopherLaughing(s Sprite) Emote {
	ms := func(n int) time.Duration { return time.Duration(n) * time.Millisecond }
	bounce := func(dy int) Transform { return Transform{DY: dy} }
	wide := append(append(Pose{}, s.Blink...), gopherMouthPose(3)...)
	return Emote{
		Name: "laughing", Loop: false,
		Frames: []Frame{
			{Pose: wide, Transform: bounce(-1), Hold: ms(140)},
			{Pose: s.Blink, Transform: bounce(0), Hold: ms(90)},
			{Pose: wide, Transform: bounce(-1), Hold: ms(140)},
			{Pose: s.Blink, Transform: bounce(0), Hold: ms(90)},
			{Pose: wide, Transform: bounce(-1), Hold: ms(160)},
			{Hold: ms(120)},
		},
	}
}

func mustGopherSprite(name string, art []string) Sprite {
	s, err := NewSprite(name, art)
	if err != nil {
		panic("auklet: built-in gopher sprite " + name + ": " + err.Error())
	}
	return s
}

// GopherSideView is the gopher in profile. Invented -- the reference is one
// front-facing sticker -- and carries no mouth: talking side-on would mean
// the same missing-mandible problem the puffin's side view has.
var GopherSideView = mustGopherSprite("side", gopherSideArt)

// GopherTurn60View and GopherTurn30View are the intermediate turn angles,
// same idea as the puffin's: eye and ear count carry the turn since there is
// no beak reach to do it here.
var GopherTurn60View = mustGopherSprite("turn60", gopherTurn60Art)
var GopherTurn30View = mustGopherSprite("turn30", gopherTurn30Art)

// GopherFrontView is the gopher head-on, matching the reference sticker.
// This is the one that can talk.
var GopherFrontView = buildGopherFront()

func buildGopherFront() Sprite {
	s, err := NewSprite("front", gopherFrontArt,
		WithMouths(
			nil, // shut: the base art's teeth, closed
			gopherMouthPose(1),
			gopherMouthPose(2),
			gopherMouthPose(3),
		),
		WithPose("wave-up", gopherWaveUp),
		WithPose("wave-tilt", gopherWaveTilt),
		WithPose("point-up", gopherWaveUp), // the same raised arm, a second name for what it means
		WithPose("point-down", gopherPointDown),
		WithPose("point-left", gopherPointLeft),
		WithPose("point-right", gopherPointRight),
		WithPose("angry", gopherAngryBrow),
		WithPose("sad", gopherSadBrow),
		WithPose("astronaut", gopherHelmet),
		WithEmote(gopherWalkIn()),
	)
	if err != nil {
		panic("auklet: built-in gopher-front sprite: " + err.Error())
	}
	// laughing needs s.Blink, which NewSprite only derives once the option
	// list above has already run.
	WithEmote(gopherLaughing(s))(&s)
	return s
}

// GopherViews is the gopher's turn set, in the same 90/60/30/0 order Views
// uses for the puffin.
func GopherViews() []Sprite {
	return []Sprite{GopherSideView, GopherTurn60View, GopherTurn30View, GopherFrontView}
}
