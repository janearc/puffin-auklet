package main

// The head-on view. Same role vocabulary as the side, so it inherits theming,
// sizing, cutout and posing without any of them learning a second view exists.
//
// The beak is laterally flattened, so head-on it is a narrow blade -- the
// feature that defines the profile is the one that nearly disappears here.
// What carries the read instead is symmetry: two cheek patches split by that
// blade, two ringed eyes set high and close.
func drawFront() *drawing {
	const W, H = 36, 46
	d := newDrawing(W, H)

	head := func(x, y int) bool { return inEllipse(x, y, 18, 14, 12, 12) }
	body := func(x, y int) bool { return inEllipse(x, y, 18, 34, 12, 10) }

	beakPoly := []pt{
		{15.4, 11}, {20.6, 11},
		{20.4, 16}, {19.7, 21}, {18.9, 26},
		{17.1, 26}, {16.3, 21}, {15.6, 16},
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
			// two cheeks, parted by where the beak lands
			if head(x, y) && y >= 8 && y <= 23 &&
				(inEllipse(x, y, 11.5, 15, 4.8, 6.5) || inEllipse(x, y, 24.5, 15, 4.8, 6.5)) {
				d.set(x, y, 'W')
			}
			// breast, leaving a black collar between it and the cheeks
			if body(x, y) && y >= 29 && inEllipse(x, y, 18, 35, 8.8, 8) {
				d.set(x, y, 'W')
			}
		}
	}

	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if !inPoly(x, y, beakPoly) {
				continue
			}
			// head-on the bands run across, so they split on y
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

	for _, cx := range []float64{11.5, 24.5} {
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				if inEllipse(x, y, cx, 13, 2.3, 2.3) {
					d.set(x, y, 'X')
				}
			}
		}
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				if inEllipse(x, y, cx, 13, 1.35, 1.35) {
					d.set(x, y, 'E')
				}
			}
		}
		for dx := -1; dx <= 1; dx++ {
			if x := int(cx) + dx; d.at(x, 10) == 'W' {
				d.set(x, 10, 'D')
			}
		}
	}

	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if body(x, y) && d.at(x, y) == 'K' &&
				(inEllipse(x, y, 7.5, 34, 3, 7.5) || inEllipse(x, y, 28.5, 34, 3, 7.5)) {
				d.set(x, y, 'V')
			}
		}
	}

	for _, fx := range []int{8, 20} {
		for y := 43; y <= 44; y++ {
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
