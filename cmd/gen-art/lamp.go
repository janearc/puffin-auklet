package main

// Luxo Jr.: a desk lamp, and the odd one out among the characters here,
// because it has no face. The real thing has never had eyes -- what reads
// as "looking at something" is the whole neck and shade tilting, and what
// reads as personality is entirely in the hop. That maps onto exactly the
// half of this package that had no test yet: Transform. Everything the
// gopher's expressions needed hand-drawn art for, the lamp gets from
// choreography instead -- see auklet/lamp.go.
//
// Distinctive, in prose, per the gen-art README's own rule: a round weighted
// base; a springy double-bent arm, thin against the base; a conical shade
// tilted forward; and the bulb glowing at the shade's mouth, which is the
// one part of this sprite that is trying to be looked AT rather than read
// as a feature -- it gets the biggest role band the same way the puffin's
// beak tip and the gopher's teeth do.
//
// Roles: K is the lamp's own plastic (arm, base, shade rim). B/Y/R are the
// shade's three bands, rim to bulb, brightest last -- the same "three bands
// of whatever protrudes" reuse the beak and the buck teeth already are. O is
// where the base actually touches the desk. No eyes: this is the first
// sprite in the package that genuinely has none, so Gaze and WideEyes find
// nothing and quietly do nothing, which is the correct behaviour for a
// subject that was never going to have a pupil.
// drawLamp is shared by both lamp views. withBall draws Luxo's other
// signature prop permanently into the art -- a POSE cannot do this: trim()
// fixes a sprite's canvas to the bounding box of whatever the BASE art
// actually paints, once, at generation time, and an overlay can only add
// pixels inside bounds that already exist. "sometimes" the ball means two
// views, front and front-ball, the same way the gopher is one character
// with several Sprites rather than one Sprite trying to be several shapes.
func drawLamp(withBall bool) *drawing {
	const W, H = 44, 48
	d := newDrawing(W, H)

	// the base: a wide low dome, weighted, flat on the desk
	base := func(x, y int) bool { return inEllipse(x, y, 17, 42, 9.5, 5) && y <= 43 }
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if base(x, y) {
				d.set(x, y, 'K')
			}
		}
	}
	// a highlight arc across the top of the base -- the one piece of
	// roundness cheap enough to afford
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if inEllipse(x, y, 15, 39, 5, 1.6) {
				d.set(x, y, 'V')
			}
		}
	}
	// no feet: tried as a full-width strip under the base and it read as a
	// second, disconnected object rather than contact. the base's own flat
	// underside already does that job -- RoleFeet stays unused here, which
	// is fine; nothing requires every sprite to use every role.

	// the arm: two straight segments, bent at an elbow, the same "springy
	// zigzag" silhouette the real lamp is known for. Poly rather than a
	// capsule, the way the puffin's beak is -- an angled thing is easier to
	// describe as a quadrilateral than as an ellipse.
	lower := []pt{
		{15.5, 39}, {19.5, 39}, {12, 21}, {8.5, 22.5},
	}
	upper := []pt{
		{9, 22}, {12.5, 20}, {23, 9.5}, {20.5, 7},
	}
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if inPoly(x, y, lower) || inPoly(x, y, upper) {
				d.set(x, y, 'K')
			}
		}
	}

	// joints: base, elbow, neck -- a small dark disc with a rivet dot, the
	// spring-hinge read
	for _, j := range []pt{{17, 39}, {10.5, 21.5}, {21.5, 8}} {
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				if inEllipse(x, y, j.x, j.y, 2.2, 2.2) {
					d.set(x, y, 'K')
				}
			}
		}
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				if inEllipse(x, y, j.x, j.y, 0.8, 0.8) {
					d.set(x, y, 'D')
				}
			}
		}
	}

	// the shade: a cone tilted forward-down, facing the viewer. three
	// bands, rim to bulb -- the bulb (R) is the biggest and brightest,
	// carrying the theme's accent, same job the puffin's beak tip has.
	shade := []pt{
		{14, 8}, {21, 4}, {28, 10}, {27, 15}, {17, 17},
	}
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if !inPoly(x, y, shade) {
				continue
			}
			switch {
			case inEllipse(x, y, 22, 12, 4.5, 4.5):
				d.set(x, y, 'R') // the bulb, glowing at the shade's mouth
			case inEllipse(x, y, 21.5, 11, 6.5, 6.5):
				d.set(x, y, 'Y')
			default:
				d.set(x, y, 'B')
			}
		}
	}

	// the ball: Luxo's other prop, resting beside the base on the same
	// ground line. Role O -- unused otherwise in this sprite (see the note
	// above the base's own feet-that-aren't) -- so it is not competing with
	// anything else in the alphabet for a colour.
	if withBall {
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				if inEllipse(x, y, 37, 41, 4.5, 4.5) {
					d.set(x, y, 'O')
				}
			}
		}
		// a highlight, the same cheap roundness trick the base's arc uses
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				if inEllipse(x, y, 35.5, 39, 1.6, 1.2) {
					d.set(x, y, 'D')
				}
			}
		}
	}

	return d
}

func drawLampFront() *drawing     { return drawLamp(false) }
func drawLampFrontBall() *drawing { return drawLamp(true) }

func lampJobs() []job {
	return []job{
		{"auklet/lamp_front_art.go", "lampFrontArt",
			"The lamp, front-on: a weighted base, a springy double-bent arm, and a tilted conical shade with the bulb glowing at its mouth. No eyes -- the real thing has never had any; what reads as looking at something is the whole neck and shade turning, which is Transform's job, not the art's.",
			drawLampFront},
		{"auklet/lamp_front_ball_art.go", "lampFrontBallArt",
			"The lamp, front-on, with the ball beside it -- Luxo's other prop. Same drawing as lampFrontArt; the ball is drawn permanently into this view rather than posed, because a pose cannot grow a sprite's canvas past what trim() already fixed at generation time.",
			drawLampFrontBall},
	}
}
