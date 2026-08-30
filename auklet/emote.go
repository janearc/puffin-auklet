package auklet

import "time"

// Emotes: named frame series, so a script can cue a performance by name.
//
// A pose is a still. An emote is a little piece of timing, and timing is most
// of what reads as feeling at this size -- a bird that blinks twice quickly and
// looks away is startled, and nothing about the drawing changed.

// A Part is a region of the bird that a pose drives. It exists so two things
// animating at once can be checked for collision instead of silently fighting:
// during narration the mouth belongs to the speech track, and an emote that
// also moves the mouth would stutter it.
type Part uint8

const (
	PartEyes Part = 1 << iota
	PartMouth
	PartOther
)

func (p Part) String() string {
	switch {
	case p == 0:
		return "none"
	case p&PartEyes != 0 && p&PartMouth != 0:
		return "eyes+mouth"
	case p&PartEyes != 0:
		return "eyes"
	case p&PartMouth != 0:
		return "mouth"
	}
	return "other"
}

// Parts reports every part this pose drives.
func (p Pose) Parts() Part {
	var out Part
	for _, o := range p {
		out |= o.Part
	}
	return out
}

// Only returns the overlays driving parts inside mask. This is how an emote is
// made safe to play during speech:
//
//	pose := emote.Pose(now).Only(^auklet.PartMouth)
//
// which keeps the eyes and drops anything that would fight the mouth track.
func (p Pose) Only(mask Part) Pose {
	var out Pose
	for _, o := range p {
		if o.Part&mask != 0 {
			out = append(out, o)
		}
	}
	return out
}

// A Transform is whole-sprite motion: a translation in cells and a size
// multiplier. The zero value is the identity, so a frame that does not mention
// it does not move.
//
// It exists because three separate things wanted it independently -- a settle
// on the rawr, a bob on the walk-on, and the anime pop -- and each was being
// hand-rolled by whoever wanted it. None of them is art: the sprite already
// resamples to any size, so a pop is rows going 11, 13, 11 and nothing is
// drawn. Putting it in Frame makes all three one mechanism, nameable in a cue
// script, and still a pure function of elapsed time.
//
// The sprite does not apply this. The CALLER does, because the caller owns the
// size and the position -- see ScaleRows. Scale is relative for the same
// reason: an emote that demands 13 rows fights a corner that only has 8, while
// one that asks for 1.2x composes with it.
type Transform struct {
	DX, DY int     // cells
	Scale  float64 // 0 means 1
}

// Factor is Scale with the zero value read as 1.
func (t Transform) Factor() float64 {
	if t.Scale == 0 {
		return 1
	}
	return t.Scale
}

// IsIdentity reports whether this transform does nothing.
func (t Transform) IsIdentity() bool {
	return t.DX == 0 && t.DY == 0 && t.Factor() == 1
}

// ScaleRows applies a factor to a row count, never returning less than 1.
func ScaleRows(rows int, factor float64) int {
	n := int(float64(rows)*factor + 0.5)
	if n < 1 {
		n = 1
	}
	return n
}

// A Frame is a pose held for a while, optionally moved or resized.
type Frame struct {
	Pose      Pose
	Hold      time.Duration
	Transform Transform
}

// An Emote is a named frame series.
//
// It settles by construction: Pose returns nil once the series is over, so a
// bird cannot be left stuck in the last frame of "startled" because the next
// cue arrived early. Interruption is likewise nothing special -- stop asking a
// performance for poses and the bird is at rest on the next frame, because a
// pose is an overlay and not a state.
type Emote struct {
	Name   string
	Frames []Frame
	Loop   bool
}

// Length is the series' total duration. A looping emote reports one cycle.
func (e Emote) Length() time.Duration {
	var d time.Duration
	for _, f := range e.Frames {
		d += f.Hold
	}
	return d
}

// Parts reports every part the series drives, so a caller can tell before
// playing whether it will collide with speech.
func (e Emote) Parts() Part {
	var out Part
	for _, f := range e.Frames {
		out |= f.Pose.Parts()
	}
	return out
}

// At returns the whole frame at elapsed time t, and whether the emote is still
// running. Use this when the emote may move or resize the bird; Pose is the
// shorthand for when it cannot.
func (e Emote) At(t time.Duration) (Frame, bool) {
	total := e.Length()
	if total <= 0 || len(e.Frames) == 0 {
		return Frame{}, false
	}
	if t < 0 {
		t = 0
	}
	if t >= total {
		if !e.Loop {
			return Frame{}, false
		}
		t %= total
	}
	var acc time.Duration
	for _, f := range e.Frames {
		acc += f.Hold
		if t < acc {
			return f, true
		}
	}
	return e.Frames[len(e.Frames)-1], true
}

// Bounds reports the extremes this emote reaches: the largest scale factor and
// the largest absolute offsets. Reserve a block at these before playing, or a
// sprite that grows mid-series will tear whatever it is spliced into.
func (e Emote) Bounds() (maxScale float64, maxAbsDX, maxAbsDY int) {
	maxScale = 1
	for _, f := range e.Frames {
		if s := f.Transform.Factor(); s > maxScale {
			maxScale = s
		}
		if d := abs(f.Transform.DX); d > maxAbsDX {
			maxAbsDX = d
		}
		if d := abs(f.Transform.DY); d > maxAbsDY {
			maxAbsDY = d
		}
	}
	return maxScale, maxAbsDX, maxAbsDY
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// Pose returns the pose at elapsed time t, and whether the emote is still
// running. A non-looping emote returns (nil, false) once it is over.
func (e Emote) Pose(t time.Duration) (Pose, bool) {
	total := e.Length()
	if total <= 0 || len(e.Frames) == 0 {
		return nil, false
	}
	if t < 0 {
		t = 0
	}
	if t >= total {
		if !e.Loop {
			return nil, false
		}
		t %= total
	}
	var acc time.Duration
	for _, f := range e.Frames {
		acc += f.Hold
		if t < acc {
			return f.Pose, true
		}
	}
	return e.Frames[len(e.Frames)-1].Pose, true
}

// Emotes lists this sprite's emotes, in the order they were attached.
func (s Sprite) Emotes() []string {
	return append([]string{}, s.emoteOrder...)
}

// Emote returns one by name. A script should validate its cues against this
// before a recording rather than discover a typo as a bird doing nothing.
func (s Sprite) Emote(name string) (Emote, bool) {
	e, ok := s.emotes[name]
	return e, ok
}

// WithEmote attaches a named frame series.
func WithEmote(e Emote) SpriteOption {
	return func(s *Sprite) {
		if s.emotes == nil {
			s.emotes = map[string]Emote{}
		}
		if _, seen := s.emotes[e.Name]; !seen {
			s.emoteOrder = append(s.emoteOrder, e.Name)
		}
		s.emotes[e.Name] = e
	}
}

// buildEmotes assembles the stock vocabulary from poses the sprite already has.
//
// None of these needed new art. That is the point worth noticing: once blink,
// gaze and the mouth exist, an emote is an arrangement of them in time, and
// timing is free. Drawing is what costs.
func buildEmotes(s *Sprite) {
	ms := func(n int) time.Duration { return time.Duration(n) * time.Millisecond }
	// a sprite with no eyes and no blink pose has nothing to be startled WITH.
	// advertising an emote whose every frame is empty would be a cue that
	// validates and then does nothing, which is worse than an unknown one.
	add := func(name string, loop bool, frames ...Frame) {
		for _, f := range frames {
			// a transform-only emote has no pose at all and is still a
			// performance, so emptiness means neither pose nor transform
			if len(f.Pose) > 0 || !f.Transform.IsIdentity() {
				WithEmote(Emote{Name: name, Frames: frames, Loop: loop})(s)
				return
			}
		}
	}

	blink := s.Blink
	look := func(dx, dy int) Pose { return s.Gaze(dx, dy) }
	// 1.6 is where it stops being a bigger eye and starts being alarm; 2.0
	// begins to read as a monocle
	wide := s.WideEyes(1.6)

	add("blink", false,
		Frame{Pose: blink, Hold: ms(120)})

	// eyes wide is the strongest thing this bird has at ANY size, and by the
	// most at the corner size where the pupil cannot travel far enough to be
	// seen. a blink shuts the eye; surprise opens it.
	add("surprised", false,
		Frame{Pose: wide, Hold: ms(360)},
		Frame{Hold: ms(80)})

	// shut, snap open, then look. the blink is the flinch and the wide eye is
	// the reaction; an earlier version was two blinks and read as a stutter.
	add("startled", false,
		Frame{Pose: blink, Hold: ms(80)},
		Frame{Pose: wide, Hold: ms(300)},
		Frame{Pose: append(append(Pose{}, wide...), look(0, -1)...), Hold: ms(200)},
		Frame{Hold: ms(100)})

	// look away, think about it, come back
	add("curious", false,
		Frame{Pose: look(1, -1), Hold: ms(400)},
		Frame{Pose: blink, Hold: ms(110)},
		Frame{Pose: look(1, -1), Hold: ms(260)},
		Frame{Hold: ms(120)})

	add("shifty", false,
		Frame{Pose: look(-1, 0), Hold: ms(260)},
		Frame{Pose: look(1, 0), Hold: ms(260)},
		Frame{Pose: look(-1, 0), Hold: ms(200)},
		Frame{Hold: ms(120)})

	add("sleepy", true,
		Frame{Pose: blink, Hold: ms(420)},
		Frame{Pose: look(0, 1), Hold: ms(700)},
		Frame{Pose: blink, Hold: ms(300)},
		Frame{Hold: ms(500)})

	add("look-away", false,
		Frame{Pose: look(-1, 1), Hold: ms(500)},
		Frame{Hold: ms(150)})

	// transforms: no art, no poses, pure timing. the caller applies them.
	pop := func(f float64) Transform { return Transform{Scale: f} }
	add("pop", false,
		Frame{Transform: pop(1.18), Hold: ms(90)},
		Frame{Transform: pop(0.94), Hold: ms(70)},
		Frame{Transform: pop(1.04), Hold: ms(60)},
		Frame{Hold: ms(40)})

	// the body drops on the attack and recovers: what makes a rawr land
	add("settle", false,
		Frame{Transform: Transform{DY: 1}, Hold: ms(70)},
		Frame{Transform: Transform{DY: 1}, Hold: ms(60)},
		Frame{Hold: ms(120)})

	add("bob", true,
		Frame{Hold: ms(180)},
		Frame{Transform: Transform{DY: -1}, Hold: ms(180)})

	add("nope", false,
		Frame{Transform: Transform{DX: -1}, Hold: ms(80)},
		Frame{Transform: Transform{DX: 1}, Hold: ms(80)},
		Frame{Transform: Transform{DX: -1}, Hold: ms(80)},
		Frame{Hold: ms(60)})

	if s.MouthLevels() > 2 {
		// the only emote that touches the mouth, and it is tagged so, so a
		// caller can refuse it during speech instead of stuttering the track
		add("gasp", false,
			Frame{Pose: append(append(Pose{}, s.Mouth(s.MouthLevels()-1)...), wide...), Hold: ms(320)},
			Frame{Pose: wide, Hold: ms(180)})
	}
}
