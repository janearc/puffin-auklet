package auklet

// Gaze: the cheap half of articulation.
//
// The beak cannot move, because nothing is drawn behind it. The pupil can,
// because the eye it sits in is already drawn -- shifting it uncovers ring, not
// a hole. That distinction is the whole of what articulation costs here: a part
// can move exactly as far as there is art behind it.
//
// So this is not a rig. It is the pupil sliding inside its own socket, which is
// enough for a bird to appear to be looking at something.

// An eye, found in the art rather than declared. Pixels carrying the pupil or
// ring roles are grouped into connected regions, so the front view's two eyes
// are two regions and the beak between them is untouched.
type eye struct {
	x0, y0, x1, y1 int             // bounds, inclusive
	socket         map[[2]int]bool // ring and pupil together: where a pupil may go
	pupil          [][2]int        // the pupil's pixels at rest
}

// eyes are found once, at construction, and carried on the sprite. an earlier
// version cached them in a package map keyed by NAME, which meant a stranger's
// sprite called "front" would have inherited this package's eye coordinates and
// put a pupil somewhere in its cheek.
func (s Sprite) eyes() []eye {
	if s.eyeCache != nil {
		return s.eyeCache
	}
	return findEyes(s.art)
}

func findEyes(art []string) []eye {
	isEye := func(x, y int) bool {
		if y < 0 || y >= len(art) || x < 0 || x >= len(art[y]) {
			return false
		}
		return art[y][x] == 'E' || art[y][x] == 'X'
	}

	seen := map[[2]int]bool{}
	var out []eye
	for y := range art {
		for x := 0; x < len(art[y]); x++ {
			if !isEye(x, y) || seen[[2]int{x, y}] {
				continue
			}
			// flood fill this region; disjoint regions are separate eyes
			e := eye{x0: x, y0: y, x1: x, y1: y, socket: map[[2]int]bool{}}
			stack := [][2]int{{x, y}}
			for len(stack) > 0 {
				p := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if seen[p] || !isEye(p[0], p[1]) {
					continue
				}
				seen[p] = true
				e.socket[p] = true
				if art[p[1]][p[0]] == 'E' {
					e.pupil = append(e.pupil, p)
				}
				if p[0] < e.x0 {
					e.x0 = p[0]
				}
				if p[0] > e.x1 {
					e.x1 = p[0]
				}
				if p[1] < e.y0 {
					e.y0 = p[1]
				}
				if p[1] > e.y1 {
					e.y1 = p[1]
				}
				stack = append(stack,
					[2]int{p[0] + 1, p[1]}, [2]int{p[0] - 1, p[1]},
					[2]int{p[0], p[1] + 1}, [2]int{p[0], p[1] - 1})
			}
			if len(e.pupil) > 0 {
				out = append(out, e)
			}
		}
	}
	return out
}

// GazeRange is how far the pupil may travel, in source pixels. One is not a
// placeholder: the pupil is two or three pixels across inside a socket barely
// wider, and at this size one pixel is already a decisive look.
const GazeRange = 1

// Gaze returns a pose with the pupils shifted by dx, dy, clamped to GazeRange.
// Gaze(0, 0) is the bird looking straight ahead and is an empty pose.
//
// Pupil pixels that would land outside the socket are dropped rather than
// painted over the face, so an extreme look squashes the pupil against the ring
// the way a real eye does.
func (s Sprite) Gaze(dx, dy int) Pose {
	dx, dy = clampGaze(dx), clampGaze(dy)
	if dx == 0 && dy == 0 {
		return nil
	}

	var pose Pose
	for _, e := range s.eyes() {
		w, h := e.x1-e.x0+1, e.y1-e.y0+1
		buf := make([][]byte, h)
		for y := range buf {
			buf[y] = make([]byte, w)
			for x := range buf[y] {
				buf[y][x] = '.' // transparent: leave the face alone
			}
		}
		// wipe the pupil back to ring, then redraw it where it is looking
		for _, p := range e.pupil {
			buf[p[1]-e.y0][p[0]-e.x0] = 'X'
		}
		for _, p := range e.pupil {
			tx, ty := p[0]+dx, p[1]+dy
			if !e.socket[[2]int{tx, ty}] {
				continue
			}
			buf[ty-e.y0][tx-e.x0] = 'E'
		}

		art := make([]string, h)
		for y := range buf {
			art[y] = string(buf[y])
		}
		pose = append(pose, Overlay{
			Name: "gaze", Part: PartEyes, OX: e.x0, OY: e.y0, Art: art,
		})
	}
	return pose
}

func clampGaze(v int) int {
	if v > GazeRange {
		return GazeRange
	}
	if v < -GazeRange {
		return -GazeRange
	}
	return v
}

// GazeToward turns a vector -- from the bird to whatever it should attend to --
// into a gaze direction. Any distance works; only the sign survives, because at
// this size there is nothing between "looking left" and "looking further left".
func GazeToward(dx, dy int) (int, int) { return sign(dx), sign(dy) }

func sign(v int) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}

// WideEyes returns a pose with the eyes opened wide by a factor.
//
// Surprise, and the cheapest legible thing this bird can do -- it beats a blink
// at every size and doubles a gaze at eight rows, which is where the corner
// bird lives. Credit for the finding to the puffin integration; scale had been
// filed under whole-sprite transforms, which is right for the bird and wrong
// for a part of it.
//
// It is legal by the same rule Gaze is: a part may move as far as there is art
// behind it, and the eye is surrounded by drawn cheek. The beak is the
// counter-example precisely because it has none.
//
// It is DRAWN rather than magnified. Scaling the socket pixel by pixel thickens
// the ring along with everything else and gives a square blob that reads as
// goggles; what reads as alarm is a thin ring at a larger radius with the pupil
// filling most of it. The pupil grows with the ring for the same reason -- a
// small pupil in a big ring reads as looking away.
//
// Growth is clipped to cheek: a pixel is painted only where the resting art is
// light or already eye, so an eye that would burst through the cap or the beak
// is trimmed rather than spilling over it. On a small face a large factor
// simply stops having an effect, which is the right failure.
//
// A face with no cheek at all -- Ms. Pac-Man's disc is one flat Dark role,
// nothing drawn Light anywhere near the eye -- has nowhere to grow into.
// Painting only the pixels already inside the resting socket is not a
// smaller version of the effect, it is a wrong one: the ring/pupil split
// gets recomputed against a radius the art never grew to match, and the
// eye comes out scrambled rather than merely unchanged. So growth counts
// only if some painted pixel actually landed on cheek; a socket that
// cannot grow returns empty for that eye, same as a face with no eye at
// all, and buildEmotes' own guard against an all-empty performance takes
// it from there.
func (s Sprite) WideEyes(factor float64) Pose {
	if factor <= 1 {
		return nil
	}
	base := s.art
	h, w := len(base), len(base[0])

	var pose Pose
	for _, e := range s.eyes() {
		cx := (float64(e.x0) + float64(e.x1) + 1) / 2
		cy := (float64(e.y0) + float64(e.y1) + 1) / 2
		rx := (float64(e.x1-e.x0+1) / 2) * factor
		ry := (float64(e.y1-e.y0+1) / 2) * factor

		x0, x1 := int(cx-rx-1), int(cx+rx+1)
		y0, y1 := int(cy-ry-1), int(cy+ry+1)
		if x0 < 0 {
			x0 = 0
		}
		if y0 < 0 {
			y0 = 0
		}
		if x1 >= w {
			x1 = w - 1
		}
		if y1 >= h {
			y1 = h - 1
		}
		if x1 < x0 || y1 < y0 {
			continue
		}

		within := func(x, y int, k float64) bool {
			dx := (float64(x) + 0.5 - cx) / (rx * k)
			dy := (float64(y) + 0.5 - cy) / (ry * k)
			return dx*dx+dy*dy <= 1
		}

		bw, bh := x1-x0+1, y1-y0+1
		buf := make([][]byte, bh)
		for y := range buf {
			buf[y] = make([]byte, bw)
			for x := range buf[y] {
				buf[y][x] = RoleNone
			}
		}

		painted, grew := false, false
		for y := y0; y <= y1; y++ {
			for x := x0; x <= x1; x++ {
				if !within(x, y, 1) {
					continue
				}
				onCheek := base[y][x] == RoleLight
				// never over the cap or the beak: cheek, or the resting eye
				if !onCheek && !e.socket[[2]int{x, y}] {
					continue
				}
				role := byte(RolePupil)
				if !within(x, y, 0.62) {
					role = RoleEyeRing // a thin ring, not a scaled band
				}
				buf[y-y0][x-x0] = role
				painted = true
				if onCheek {
					grew = true
				}
			}
		}
		if !painted || !grew {
			continue
		}

		art := make([]string, bh)
		for y := range buf {
			art[y] = string(buf[y])
		}
		pose = append(pose, Overlay{
			Name: "wide-eyes", Part: PartEyes, OX: x0, OY: y0, Art: art,
		})
	}
	return pose
}

// derivedBlink builds a lid from the art, for a sprite that did not declare
// one. A hand-drawn blink is better -- the built-in views have theirs -- but a
// derived one means Views() has no holes and a stranger's sprite can shut its
// eyes without knowing the format has a `pose blink` section.
func (s Sprite) derivedBlink() Pose {
	var pose Pose
	for _, e := range s.eyes() {
		w, h := e.x1-e.x0+1, e.y1-e.y0+1
		lid := e.y0 + h/2
		art := make([]string, h)
		for y := 0; y < h; y++ {
			row := make([]byte, w)
			for x := 0; x < w; x++ {
				switch {
				case !e.socket[[2]int{e.x0 + x, e.y0 + y}]:
					row[x] = RoleNone
				case e.y0+y == lid:
					row[x] = RoleStripe
				default:
					row[x] = RoleLight
				}
			}
			art[y] = string(row)
		}
		pose = append(pose, Overlay{
			Name: "blink", Part: PartEyes, OX: e.x0, OY: e.y0, Art: art,
		})
	}
	return pose
}
