package auklet

import "time"

// Expressions and a costume for the puffin's FrontView, attached the same
// way a stranger's sprite would attach them -- WithPose and WithEmote,
// mutating the built-in global after the fact. Nothing here touches
// sprite.go, pose.go or any *_art.go file.
//
// The puffin has less to work with than the gopher and it shows: no limbs
// exist anywhere in its art, so there is nothing to point with, and the
// brow read is genuinely faint -- one extra pixel per eye is what two rows
// of spare canvas above a 5-pixel-wide eye affords. Honest about the limit
// rather than inventing head-canon anatomy to work around it.

// puffinAngryBrow, puffinSadBrow: a one-row tilt above the eye didn't read
// at this resolution on the gopher either -- vertical position and weight
// carry further than angle does at a five-pixel eye. Angry thickens the
// existing plate mark (row 8) into a solid bar at the eye's own edge; sad
// moves a thinner one three rows further up, onto the forehead, well clear
// of the eye. The gopher's version of this note explains the reasoning;
// this is the same fix, applied to less canvas.
var puffinAngryBrow = Pose{
	{Name: "brow-l", Part: PartEyes, OX: 3, OY: 8, Art: []string{"DDDDD"}},
	{Name: "brow-r", Part: PartEyes, OX: 16, OY: 8, Art: []string{"DDDDD"}},
}

var puffinSadBrow = Pose{
	{Name: "brow-l", Part: PartEyes, OX: 3, OY: 5, Art: []string{"DDD"}},
	{Name: "brow-r", Part: PartEyes, OX: 18, OY: 5, Art: []string{"DDD"}},
}

// puffinHelmet: a full ring computed from the head ellipse drawFront uses
// (center 12,12 trimmed, radius 12 -- see cmd/gen-art/front.go), the same
// way the gopher's is computed rather than eyeballed.
var puffinHelmet = Pose{{
	Name: "astronaut", Part: PartOther, OX: -2, OY: -2,
	Art: []string{
		"..........XXXXXXXX..........",
		"........XXXXXXXXXXXX........",
		"......XXXX........XXXX......",
		".....XXX............XXX.....",
		"....XX................XX....",
		"...XX..................XX...",
		"..XX....................XX..",
		"..XX....................XX..",
		".XX......................XX.",
		".XX......................XX.",
		"XX........................XX",
		"XX........................XX",
		"XX........................XX",
		"XX........................XX",
		"XX........................XX",
		"XX........................XX",
		"XX........................XX",
		"XX........................XX",
		".XX......................XX.",
		".XX......................XX.",
		"..XX....................XX..",
		"..XX....................XX..",
		"...XX..................XX...",
		"....XX................XX....",
		".....XXX............XXX.....",
		"......XXXX........XXXX......",
		"........XXXXXXXXXXXX........",
		"..........XXXXXXXX..........",
	},
}}

// Turn30View's mouth. Only FrontView could talk before this -- Side and the
// two turn views all draw the beak as a length-wise wedge or a diagonal
// band, which has nothing to grow into without inventing a mandible from
// nothing. Turn30 is the one exception: drawTurn30 (cmd/gen-art/turn.go)
// bands its beak by ROW, the same flattened-blade style front.go uses, just
// off-centre. Coordinates read off auklet/turn30_art.go directly rather
// than eyeballed: the gumline (Y) sits at rows 13-14, cols 5-10; the tip (R)
// starts at row 15 and narrows to cols 6-8 by row 20, ending at row 23.
//
// The technique is FrontView's Mouths, transplanted: repaint the gumline
// and however much of the tip an openness level calls for as dark, so the
// narrower tip visible below reads as a mouth that has opened; "wide" also
// extends two more rows of tip past where the base art's beak actually
// ends, the same "wide-jaw" trick front_art.go's own widest level uses.
var puffinTurn30Mouths = []Pose{
	nil, // shut: the base art's beak, closed
	{{Name: "mouth", Part: PartMouth, OX: 5, OY: 13, Art: []string{"KKKKKK", "KKKKKK"}}},
	{{Name: "mouth", Part: PartMouth, OX: 5, OY: 13, Art: []string{
		"KKKKKK", "KKKKKK", "KKKKKK", "KKKKKK",
	}}},
	{
		{Name: "mouth", Part: PartMouth, OX: 5, OY: 12, Art: []string{
			"KKKKKK", "KKKKKK", "KKKKKK", "KKKKKK", "KKKKKK", "KKKKKK", "KKKKKK",
		}},
		{Name: "wide-jaw", Part: PartMouth, OX: 6, OY: 24, Art: []string{"RRR", "RRR"}},
	},
}

func init() {
	WithMouths(puffinTurn30Mouths...)(&Turn30View)

	WithPose("angry", puffinAngryBrow)(&FrontView)
	WithPose("sad", puffinSadBrow)(&FrontView)
	WithPose("astronaut", puffinHelmet)(&FrontView)

	ms := func(n int) time.Duration { return time.Duration(n) * time.Millisecond }
	bounce := func(dy int) Transform { return Transform{DY: dy} }
	wide := append(append(Pose{}, FrontView.Blink...), FrontView.Mouth(FrontView.MouthLevels()-1)...)
	WithEmote(Emote{
		Name: "laughing", Loop: false,
		Frames: []Frame{
			{Pose: wide, Transform: bounce(-1), Hold: ms(140)},
			{Pose: FrontView.Blink, Transform: bounce(0), Hold: ms(90)},
			{Pose: wide, Transform: bounce(-1), Hold: ms(140)},
			{Pose: FrontView.Blink, Transform: bounce(0), Hold: ms(90)},
			{Pose: wide, Transform: bounce(-1), Hold: ms(160)},
			{Hold: ms(120)},
		},
	})(&FrontView)
}
