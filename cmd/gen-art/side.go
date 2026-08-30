package main

// The side view. Read this top to bottom and it is the order the bird was
// actually drawn in, which is also the order that matters:
//
//  1. silhouette solid, head and body as one shape
//  2. white face and belly carved OUT of it -- painting white on black instead
//     loses the throat collar every time, and the collar is half the read
//  3. the beak, banded ACROSS its own axis rather than straight down; bands
//     parallel to x read as a traffic cone
//  4. the eye, small, then the plate and nape stripe
//  5. wing tone, then feet
func drawSide() *drawing {
	const W, H = 48, 48
	d := newDrawing(W, H)

	head := func(x, y int) bool { return inEllipse(x, y, 25, 14, 11.5, 11) }
	body := func(x, y int) bool { return inEllipse(x, y, 27, 31, 14.5, 12.5) }

	// the beak's base is nearly as deep as the head is tall. that one
	// proportion is most of what makes it a puffin rather than a seabird.
	beakPoly := []pt{
		{2, 20.5}, {4, 17}, {7, 13.5}, {10.5, 10.5}, {14, 8}, {17.5, 6.5},
		{19, 12}, {19, 18}, {18, 23}, {16.5, 27},
		{12, 26}, {8, 24.5}, {4.5, 22.5},
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
			if body(x, y) && y >= 28 && inEllipse(x, y, 24.5, 34, 11.5, 10) {
				d.set(x, y, 'W')
			}
			if head(x, y) && y >= 8 && y <= 22 && inEllipse(x, y, 21.5, 14.5, 8.5, 8) {
				d.set(x, y, 'W')
			}
		}
	}

	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if !beak(x, y) {
				continue
			}
			// u runs along the beak, so the bands sit across it
			u := float64(x) - 0.18*(float64(y)-20)
			switch {
			case u >= 14:
				d.set(x, y, 'B')
			case u >= 12:
				d.set(x, y, 'Y')
			default:
				d.set(x, y, 'R')
			}
		}
	}
	// the gape line along the beak's lower edge
	for x := 3; x <= 17; x++ {
		for y := H - 1; y >= 0; y-- {
			if beak(x, y) {
				d.set(x, y, 'Y')
				break
			}
		}
	}

	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if inEllipse(x, y, 22, 12.5, 2.4, 2.2) {
				d.set(x, y, 'X')
			}
		}
	}
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if inEllipse(x, y, 22, 12.5, 1.6, 1.5) {
				d.set(x, y, 'E')
			}
		}
	}
	for x := 20; x <= 23; x++ {
		d.set(x, 10, 'D')
	}
	// the nape stripe stops where the white face does, or it stains the cap
	for i := 0; i < 7; i++ {
		x, y := 25+i, 13+i/3
		if d.at(x, y) == 'W' {
			d.set(x, y, 'D')
		}
	}

	// just enough tone to stop the back reading as a flat blob
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if body(x, y) && d.at(x, y) == 'K' && inEllipse(x, y, 35, 29, 4.5, 7) {
				d.set(x, y, 'V')
			}
		}
	}

	for _, fx := range []int{19, 28} {
		for y := 41; y <= 43; y++ {
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
	return d
}
