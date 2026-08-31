package main

// Ms. Pac-Man: a circle, a bow, and a mouth that is the character's whole
// personality. Distinctive, per the gen-art README's own rule: the yellow
// disc; the bow (two lobes and a knot, the one thing that says "Ms." rather
// than the plain arcade original); a single eye, which the real sprite
// rarely bothers with but this package leans on hard -- Gaze, WideEyes and
// half the stock emote vocabulary all key off an eye existing at all; the
// beauty mark; and the wedge mouth, drawn SHUT here on purpose -- see
// auklet/mspacman.go for why the open frames are overlays, not more art.
//
// Roles: K is the disc itself -- the one flat mass, same slot the gopher's
// body and the lamp's chassis already used it for. B/Y/R are the bow's rim,
// shade and accent, the same "three bands of whatever protrudes" reuse the
// beak, the buck teeth and the shade already are. O is a pair of small red
// shoes at the bottom -- the 80s-cartoon touch, and RoleFeet actually
// meaning feet for once.
func drawMsPacmanFront() *drawing {
	const W, H = 34, 40
	d := newDrawing(W, H)
	const cx, cy, r = 17.0, 23.0, 13.0

	body := func(x, y int) bool { return inEllipse(x, y, cx, cy, r, r) }
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if body(x, y) {
				d.set(x, y, 'K')
			}
		}
	}

	// the highlight: a crescent of shine, upper-left, the same cheap
	// roundness trick the gopher's base and the lamp's chassis both used.
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if body(x, y) && inEllipse(x, y, cx-5, cy-6, 4, 5) {
				d.set(x, y, 'V')
			}
		}
	}

	// the eye: one, not two -- the real sprite mostly has none at all, and
	// one is already a choice made for this package's sake, not the
	// character's. Set back from the mouth side (left), the way a face
	// looking the direction it is about to move keeps its eye trailing.
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if inEllipse(x, y, cx-4, cy-6, 2.1, 2.3) {
				d.set(x, y, 'X')
			}
		}
	}
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if inEllipse(x, y, cx-4, cy-6, 1.1, 1.2) {
				d.set(x, y, 'E')
			}
		}
	}

	// the beauty mark: one dark pixel, low and to the mouth side -- the
	// detail that says Ms. as much as the bow does.
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if inEllipse(x, y, cx+7, cy+7, 0.9, 0.9) {
				d.set(x, y, 'D')
			}
		}
	}

	// the bow: two lobes and a knot, sitting on top of the disc. The knot
	// carries the accent (R, same job the beak tip and the bulb already
	// have); the lobes shade rim to fill (B outer, Y inner).
	for _, lobe := range []struct{ cx, cy float64 }{{cx - 6, cy - 15}, {cx + 6, cy - 15}} {
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				if inEllipse(x, y, lobe.cx, lobe.cy, 5, 4) {
					d.set(x, y, 'B')
				}
			}
		}
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				if inEllipse(x, y, lobe.cx, lobe.cy, 3, 2.4) {
					d.set(x, y, 'Y')
				}
			}
		}
	}
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if inEllipse(x, y, cx, cy-14, 2.6, 2.6) {
				d.set(x, y, 'R')
			}
		}
	}

	// the shoes: two small marks at the bottom, the 80s-cartoon touch.
	// RoleFeet, actually meaning feet.
	for _, fx := range []float64{cx - 5, cx + 5} {
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				if inEllipse(x, y, fx, cy+r-1.5, 2.4, 1.6) {
					d.set(x, y, 'O')
				}
			}
		}
	}

	return d
}

func msPacmanJobs() []job {
	return []job{
		{"auklet/mspacman_front_art.go", "msPacmanFrontArt",
			"Ms. Pac-Man, mouth shut: a yellow disc, a bow, one eye, a beauty mark, and two small shoes. The wedge mouth is not drawn here at all -- see auklet/mspacman.go for why the open frames are overlays.",
			drawMsPacmanFront},
	}
}
