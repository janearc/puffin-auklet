package main

// The Go gopher. A mascot, not an animal, and the read is nothing like the
// puffin's: there is no beak to drive proportion or the turn. The signature
// is two enormous white eyes with black pupils, a black nose, two white
// buck teeth on a tan muzzle, small round ears, and a plain teal capsule
// body with tan paws and feet.
//
// Role reuse, same as the gen-art README's own example: the buck teeth are
// BeakBase/BeakBand/BeakTip (three bands of a thing that sticks out of the
// face, same as the puffin's beak) and the paws and feet are Feet. Nothing
// new was added to the role alphabet.
//
// The reference is a single front-facing sticker. Side, turn60 and turn30
// below are invented rather than traced -- there is no profile to copy from
// -- and use eye/ear count to carry the turn the way the puffin's beak reach
// does: one eye and one ear in profile, both fully in by front-on.

// capsule is a stadium: rounded top, straight sides, rounded bottom. It is
// the gopher's silhouette at every angle; only the width and where the caps
// fall change between views.
func capsule(cx, topY, botY, rx, headRy, footRy float64) func(x, y int) bool {
	return func(x, y int) bool {
		fx, fy := float64(x)+0.5, float64(y)+0.5
		switch {
		case fy < topY:
			return inEllipse(x, y, cx, topY, rx, headRy)
		case fy > botY:
			return inEllipse(x, y, cx, botY, rx, footRy)
		default:
			return fx >= cx-rx && fx <= cx+rx
		}
	}
}

// gopherEye stamps one huge eye: white sclera clipped to the body silhouette
// -- the sticker leaves only a sliver of head around it -- a ring, then the
// pupil. Same nesting the puffin's front view uses, just scaled up because
// here the white IS the eye rather than a face patch it sits in.
func gopherEye(d *drawing, body func(x, y int) bool, cx, cy, rx, ry float64) {
	for y := 0; y < d.h; y++ {
		for x := 0; x < d.w; x++ {
			if body(x, y) && inEllipse(x, y, cx, cy, rx, ry) {
				d.set(x, y, 'W')
			}
		}
	}
	for y := 0; y < d.h; y++ {
		for x := 0; x < d.w; x++ {
			if inEllipse(x, y, cx, cy, rx*0.52, ry*0.52) {
				d.set(x, y, 'X')
			}
		}
	}
	for y := 0; y < d.h; y++ {
		for x := 0; x < d.w; x++ {
			if inEllipse(x, y, cx, cy, rx*0.30, ry*0.30) {
				d.set(x, y, 'E')
			}
		}
	}
}

// gopherMuzzle stamps the tan snout patch, gumline and two teeth, centred at
// cx. It is drawn last so nothing else bleeds into it.
func gopherMuzzle(d *drawing, cx, topY float64) {
	for y := 0; y < d.h; y++ {
		for x := 0; x < d.w; x++ {
			if inEllipse(x, y, cx, topY+3.4, 4.4, 3.8) && float64(y) >= topY {
				d.set(x, y, 'B')
			}
		}
	}
	gum := int(topY) + 4
	for x := int(cx) - 4; x <= int(cx)+4; x++ {
		d.set(x, gum, 'Y')
	}
	for _, tx := range []int{int(cx) - 2, int(cx) + 1} {
		for y := gum + 1; y <= gum+4; y++ {
			for x := tx; x <= tx+1; x++ {
				d.set(x, y, 'R')
			}
		}
	}
	// the nose sits just above the muzzle, between the eyes
	for y := 0; y < d.h; y++ {
		for x := 0; x < d.w; x++ {
			if inEllipse(x, y, cx, topY-1.6, 1.8, 1.4) {
				d.set(x, y, 'D')
			}
		}
	}
}

func gopherEar(d *drawing, cx, cy float64) {
	for y := 0; y < d.h; y++ {
		for x := 0; x < d.w; x++ {
			if inEllipse(x, y, cx, cy, 3.2, 3.8) {
				d.set(x, y, 'K')
			}
		}
	}
	for y := 0; y < d.h; y++ {
		for x := 0; x < d.w; x++ {
			if inEllipse(x, y, cx, cy, 1.1, 1.3) {
				d.set(x, y, 'D')
			}
		}
	}
}

func gopherFeet(d *drawing, y0 int, xs ...int) {
	for _, fx := range xs {
		for y := y0; y < d.h; y++ {
			for x := fx; x <= fx+6; x++ {
				d.set(x, y, 'O')
			}
		}
	}
}

// gopherShade adds a sliver of darker tone down each flank so the body does
// not read as a flat teal slab. Decorative, same job RoleWing does for the
// puffin.
func gopherShade(d *drawing, body func(x, y int) bool, left, right, cy, ry float64) {
	for y := 0; y < d.h; y++ {
		for x := 0; x < d.w; x++ {
			if body(x, y) && d.at(x, y) == 'K' &&
				(inEllipse(x, y, left, cy, 2.2, ry) || inEllipse(x, y, right, cy, 2.2, ry)) {
				d.set(x, y, 'V')
			}
		}
	}
}

// drawGopherFront: head-on, matching the reference sticker. Both eyes and
// both ears at full size and dead symmetric.
func drawGopherFront() *drawing {
	const W, H = 32, 46
	d := newDrawing(W, H)
	const cx = 16.0
	body := capsule(cx, 11, 34, 10, 10, 9)

	gopherEar(d, 5.5, 9)
	gopherEar(d, 26.5, 9)
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if body(x, y) {
				d.set(x, y, 'K')
			}
		}
	}
	// ears were drawn first and the body silhouette pass just painted over
	// the part of them it overlaps; redraw the ears so they still show
	// where they stick out past the head.
	gopherEar(d, 5.5, 9)
	gopherEar(d, 26.5, 9)

	gopherEye(d, body, 9.4, 16, 5.6, 6.4)
	gopherEye(d, body, 22.6, 16, 5.6, 6.4)
	gopherMuzzle(d, cx, 24)

	// arms: short stubs at mid-body, reaching past the silhouette
	for _, ax := range []int{1, 27} {
		for y := 36; y <= 39; y++ {
			for x := ax; x <= ax+4; x++ {
				d.set(x, y, 'O')
			}
		}
	}
	gopherFeet(d, 43, 8, 17)
	gopherShade(d, body, 7.5, 24.5, 30, 8)
	return d
}

// drawGopherSide: invented. One ear, one eye, the muzzle nudged toward the
// front of the face rather than centred, and the far arm hidden.
func drawGopherSide() *drawing {
	const W, H = 30, 46
	d := newDrawing(W, H)
	const cx = 15.0
	body := capsule(cx, 11, 34, 9.5, 9.5, 8.5)

	gopherEar(d, 12, 8)
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if body(x, y) {
				d.set(x, y, 'K')
			}
		}
	}
	gopherEar(d, 12, 8)

	// the eye sits forward, toward the muzzle, the way a profile eye does
	gopherEye(d, body, 19, 16, 5.8, 6.6)
	gopherMuzzle(d, 21, 25)

	// one arm, the near one
	for y := 36; y <= 39; y++ {
		for x := 22; x <= 26; x++ {
			d.set(x, y, 'O')
		}
	}
	gopherFeet(d, 43, 6, 15)
	gopherShade(d, body, 6.5, 23.5, 30, 8)
	return d
}

// drawGopherTurn60: mostly profile. The far ear starts to clear the head's
// silhouette and the far eye is a bare hint at the edge -- not a second full
// eye, which would read as front-on early.
func drawGopherTurn60() *drawing {
	const W, H = 31, 46
	d := newDrawing(W, H)
	const cx = 15.5
	body := capsule(cx, 11, 34, 9.7, 9.7, 8.6)

	gopherEar(d, 12.5, 8)
	gopherEar(d, 21, 8.6)
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if body(x, y) {
				d.set(x, y, 'K')
			}
		}
	}
	gopherEar(d, 12.5, 8)
	gopherEar(d, 21, 8.6)

	gopherEye(d, body, 18.6, 16, 5.9, 6.6)
	// the far eye: a narrow sliver near the head's edge
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if body(x, y) && inEllipse(x, y, 27, 16.5, 2, 5.4) {
				d.set(x, y, 'W')
			}
		}
	}
	gopherMuzzle(d, 20.5, 25)

	for y := 36; y <= 39; y++ {
		for x := 22; x <= 26; x++ {
			d.set(x, y, 'O')
		}
	}
	gopherFeet(d, 43, 6, 16)
	gopherShade(d, body, 6.5, 24, 30, 8)
	return d
}

// drawGopherTurn30: mostly front. Both eyes show, the far one compressed
// toward the edge by the turn, both ears in but not yet symmetric.
func drawGopherTurn30() *drawing {
	const W, H = 32, 46
	d := newDrawing(W, H)
	const cx = 16.5
	body := capsule(cx, 11, 34, 10, 10, 9)

	gopherEar(d, 7, 9)
	gopherEar(d, 25.5, 8.6)
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if body(x, y) {
				d.set(x, y, 'K')
			}
		}
	}
	gopherEar(d, 7, 9)
	gopherEar(d, 25.5, 8.6)

	gopherEye(d, body, 11, 16, 5.9, 6.6)
	gopherEye(d, body, 23, 16, 4.6, 6.2) // far eye, squeezed
	gopherMuzzle(d, 17.5, 24)

	for _, ax := range []int{2, 27} {
		for y := 36; y <= 39; y++ {
			for x := ax; x <= ax+4; x++ {
				d.set(x, y, 'O')
			}
		}
	}
	gopherFeet(d, 43, 8, 18)
	gopherShade(d, body, 8, 25.5, 30, 8)
	return d
}

// gopherJobs is appended into main's job table rather than inlined there, so
// adding the gopher touches main.go by exactly one line.
func gopherJobs() []job {
	return []job{
		{"auklet/gopher_side_art.go", "gopherSideArt",
			"The gopher in profile. Invented -- the reference is a single front-facing sticker -- and carries the turn on one eye and one ear rather than a beak.",
			drawGopherSide},
		{"auklet/gopher_turn60_art.go", "gopherTurn60Art",
			"The gopher 60 degrees off head-on: mostly profile, with the far ear clearing the silhouette and the far eye a bare sliver.",
			drawGopherTurn60},
		{"auklet/gopher_turn30_art.go", "gopherTurn30Art",
			"The gopher 30 degrees off head-on: mostly front, the far eye compressed toward the edge.",
			drawGopherTurn30},
		{"auklet/gopher_front_art.go", "gopherFrontArt",
			"The gopher head-on, matching the reference sticker: two huge symmetric eyes, buck teeth on a tan muzzle, round ears, plain capsule body.",
			drawGopherFront},
	}
}
