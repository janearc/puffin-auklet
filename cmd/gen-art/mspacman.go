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
// drawMsPacmanAt draws the disc with every feature -- highlight, eye, beauty
// mark, bow, shoes -- rotated rigidly by thetaDeg around the disc's own
// center. The disc's own fill is a perfect circle and needs no rotation; a
// circle is the one shape here where "spin the character" and "spin the
// drawing" are different things, and this only does the second: nothing here
// draws a new silhouette, it just moves where the features sit on the one
// silhouette that already exists.
func drawMsPacmanAt(thetaDeg float64) *drawing {
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

	// paintRot paints one feature's ellipse, its center orbited around
	// (cx,cy) by thetaDeg and its own axes turned to match -- a bow lobe
	// keeps pointing the way it is orbiting instead of staying
	// axis-aligned while only its position moves.
	paintRot := func(fx, fy, rx, ry float64, role byte, clipToBody bool) {
		nx, ny := rotatePoint(fx, fy, cx, cy, thetaDeg)
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				if clipToBody && !body(x, y) {
					continue
				}
				if inEllipseRot(x, y, nx, ny, rx, ry, thetaDeg) {
					d.set(x, y, role)
				}
			}
		}
	}

	// the highlight: a crescent of shine, upper-left, the same cheap
	// roundness trick the gopher's base and the lamp's chassis both used.
	paintRot(cx-5, cy-6, 4, 5, 'V', true)

	// the eye: one, not two -- the real sprite mostly has none at all, and
	// one is already a choice made for this package's sake, not the
	// character's. Set back from the mouth side (left), the way a face
	// looking the direction it is about to move keeps its eye trailing.
	paintRot(cx-4, cy-6, 2.1, 2.3, 'X', false)
	paintRot(cx-4, cy-6, 1.1, 1.2, 'E', false)

	// the beauty mark: one dark pixel, low and to the mouth side -- the
	// detail that says Ms. as much as the bow does.
	paintRot(cx+7, cy+7, 0.9, 0.9, 'D', false)

	// the bow: two lobes and a knot, sitting on top of the disc. The knot
	// carries the accent (R, same job the beak tip and the bulb already
	// have); the lobes shade rim to fill (B outer, Y inner).
	for _, lobe := range []struct{ cx, cy float64 }{{cx - 6, cy - 15}, {cx + 6, cy - 15}} {
		paintRot(lobe.cx, lobe.cy, 5, 4, 'B', false)
		paintRot(lobe.cx, lobe.cy, 3, 2.4, 'Y', false)
	}
	paintRot(cx, cy-14, 2.6, 2.6, 'R', false)

	// the shoes: two small marks at the bottom, the 80s-cartoon touch.
	// RoleFeet, actually meaning feet.
	for _, fx := range []float64{cx - 5, cx + 5} {
		paintRot(fx, cy+r-1.5, 2.4, 1.6, 'O', false)
	}

	return d
}

func drawMsPacmanFront() *drawing { return drawMsPacmanAt(0) }

// msPacmanSpinFrames renders drawMsPacmanAt at n evenly spaced angles and
// crops every one to the SAME absolute window drawMsPacmanFront trims to --
// not re-trimmed per frame, because a frame whose bow happens to poke out a
// pixel further at some angle would then get its own slightly different
// bounding box, and every frame has to share one registration to be laid
// down as a full-canvas overlay at OX=OY=0 without the disc appearing to
// drift frame to frame.
func msPacmanSpinFrames(n int) [][]string {
	rows0, dx0, dy0 := drawMsPacmanFront().trim()
	h0, w0 := len(rows0), len(rows0[0])

	out := make([][]string, n)
	for i := 0; i < n; i++ {
		theta := float64(i) * 360 / float64(n)
		d := drawMsPacmanAt(theta)
		frame := make([]string, h0)
		for y := 0; y < h0; y++ {
			row := make([]byte, w0)
			for x := 0; x < w0; x++ {
				row[x] = d.at(dx0+x, dy0+y)
			}
			frame[y] = string(row)
		}
		out[i] = frame
	}
	return out
}

func msPacmanJobs() []job {
	return []job{
		{"auklet/mspacman_front_art.go", "msPacmanFrontArt",
			"Ms. Pac-Man, mouth shut: a yellow disc, a bow, one eye, a beauty mark, and two small shoes. The wedge mouth is not drawn here at all -- see auklet/mspacman.go for why the open frames are overlays.",
			drawMsPacmanFront},
	}
}
