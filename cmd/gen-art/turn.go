package main

// The turn: profile to fourth wall, for the opening shot.
//
// Four orientations, 90 / 60 / 30 / 0 degrees, where 90 is full profile facing
// left and 0 is head-on. 90 and 0 are the existing side and front views; these
// are the two in between.
//
// The beak drives the whole thing. It is a laterally flattened blade -- long
// forward, tall, thin across -- so in profile you see its LENGTH and head-on
// you see its WIDTH. Everything else follows from that: as the bird turns, the
// beak's horizontal reach collapses, the far cheek comes into view around it,
// and the far eye appears last.
//
// The frames are drawn rather than interpolated. Interpolating between two
// shapes that differ qualitatively -- a wedge and a blade -- gives a shape that
// is neither, and at this size a shape that is nearly right is wrong.

// drawTurn60 is mostly profile, just off it. The tell is the far cheek starting
// to show past the beak; the far eye is still hidden behind it.
func drawTurn60() *drawing {
	const W, H = 44, 48
	d := newDrawing(W, H)

	head := func(x, y int) bool { return inEllipse(x, y, 22, 14, 12, 11.5) }
	body := func(x, y int) bool { return inEllipse(x, y, 23, 32, 13, 12.5) }

	// foreshortened: the tip has swung toward us, so it reaches less far left
	beakPoly := []pt{
		{8, 20}, {10, 17}, {12.5, 13.5}, {15, 10.5}, {17.5, 9},
		{19.5, 12.5}, {19.5, 18}, {18.5, 23}, {17, 26},
		{13.5, 25}, {11, 23.5}, {9, 22},
	}
	beak := func(x, y int) bool { return inPoly(x, y, beakPoly) }

	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if head(x, y) || body(x, y) {
				d.set(x, y, 'K')
			}
		}
	}
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if body(x, y) && y >= 28 && inEllipse(x, y, 21.5, 35, 10.5, 10) {
				d.set(x, y, 'W')
			}
			if head(x, y) && y >= 8 && y <= 23 {
				// the near cheek, and the far one appearing around the beak:
				// that second patch is what says "turned" rather than "profile"
				// no far cheek here: at this angle it would be a sliver, and a
				// sliver separated from the near cheek by black reads as a
				// floating object rather than as the far side of a face. the
				// turn is carried by the beak's reach and a rounder head.
				if inEllipse(x, y, 20, 15, 7.2, 7.5) {
					d.set(x, y, 'W')
				}
			}
		}
	}
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if !beak(x, y) {
				continue
			}
			u := float64(x) - 0.18*(float64(y)-20)
			switch {
			case u >= 15:
				d.set(x, y, 'B')
			case u >= 13:
				d.set(x, y, 'Y')
			default:
				d.set(x, y, 'R')
			}
		}
	}
	for x := 9; x <= 18; x++ {
		for y := H - 1; y >= 0; y-- {
			if beak(x, y) {
				d.set(x, y, 'Y')
				break
			}
		}
	}

	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if inEllipse(x, y, 21.5, 12.5, 2.4, 2.2) {
				d.set(x, y, 'X')
			}
		}
	}
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if inEllipse(x, y, 21.5, 12.5, 1.6, 1.5) {
				d.set(x, y, 'E')
			}
		}
	}
	for x := 20; x <= 23; x++ {
		d.set(x, 10, 'D')
	}
	for i := 0; i < 5; i++ {
		if x, y := 25+i, 13+i/3; d.at(x, y) == 'W' {
			d.set(x, y, 'D')
		}
	}

	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if body(x, y) && d.at(x, y) == 'K' && inEllipse(x, y, 32, 30, 4, 7) {
				d.set(x, y, 'V')
			}
		}
	}
	feet(d, 16, 25)
	return d
}

// drawTurn30 is mostly front, just off it. Both eyes show, the far one squeezed
// toward the edge, and the beak is nearly a blade but still swung left.
func drawTurn30() *drawing {
	const W, H = 40, 48
	d := newDrawing(W, H)

	head := func(x, y int) bool { return inEllipse(x, y, 20, 14, 12.5, 12) }
	body := func(x, y int) bool { return inEllipse(x, y, 20, 33, 12.5, 11) }

	// almost end-on: mostly width, a little length still showing to the left
	beakPoly := []pt{
		{14.2, 11}, {20.0, 11},
		{19.2, 16}, {17.8, 21}, {16.4, 26},
		{14.4, 26}, {13.4, 21}, {13.4, 16},
	}

	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if head(x, y) || body(x, y) {
				d.set(x, y, 'K')
			}
		}
	}
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if head(x, y) && y >= 8 && y <= 23 &&
				// near cheek full, far cheek narrowed by the turn
				(inEllipse(x, y, 11.5, 15, 5.2, 6.5) || inEllipse(x, y, 24.3, 15, 4.2, 6.2)) {
				d.set(x, y, 'W')
			}
			if body(x, y) && y >= 29 && inEllipse(x, y, 19.5, 35, 9, 8) {
				d.set(x, y, 'W')
			}
		}
	}
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if !inPoly(x, y, beakPoly) {
				continue
			}
			switch {
			case y < 15:
				d.set(x, y, 'B')
			case y < 17:
				d.set(x, y, 'Y')
			default:
				d.set(x, y, 'R')
			}
		}
	}

	// near eye at rest size, far eye compressed: the asymmetry IS the turn
	for _, e := range []struct{ cx, rx float64 }{{11.5, 2.3}, {24.3, 1.9}} {
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				if inEllipse(x, y, e.cx, 13, e.rx, 2.3) {
					d.set(x, y, 'X')
				}
			}
		}
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				if inEllipse(x, y, e.cx, 13, e.rx*0.6, 1.35) {
					d.set(x, y, 'E')
				}
			}
		}
		for dx := -1; dx <= 1; dx++ {
			if x := int(e.cx) + dx; d.at(x, 10) == 'W' {
				d.set(x, 10, 'D')
			}
		}
	}

	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if body(x, y) && d.at(x, y) == 'K' &&
				(inEllipse(x, y, 9, 33, 3, 7) || inEllipse(x, y, 30.5, 33, 3, 7)) {
				d.set(x, y, 'V')
			}
		}
	}
	feet(d, 11, 22)
	return d
}

// feet draws two of them, left toe at each x.
func feet(d *drawing, xs ...int) {
	for _, fx := range xs {
		for y := 42; y <= 43; y++ {
			for x := fx + 3; x <= fx+4; x++ {
				d.set(x, y, 'O')
			}
		}
		for x := fx; x <= fx+8; x++ {
			d.set(x, 44, 'O')
		}
		for x := fx + 1; x <= fx+7; x++ {
			d.set(x, 45, 'O')
		}
	}
}
