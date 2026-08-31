package auklet

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/janearc/puffin-auklet/canvas"
)

func TestCellsScale(t *testing.T) {
	th := DefaultTheme()
	for _, s := range []int{1, 2, 3, 4} {
		rows := Cells(th, s)
		wantW, wantH := SizeAt(s)
		if len(rows) != wantH {
			t.Errorf("scale %d: got %d rows, want %d", s, len(rows), wantH)
		}
		if len(rows[0]) != wantW {
			t.Errorf("scale %d: got %d cols, want %d", s, len(rows[0]), wantW)
		}
	}
}

func TestCutoutLeavesEdges(t *testing.T) {
	th := DefaultTheme()
	th.Background = nil
	rows := Cells(th, 1)

	var empty, half int
	for _, row := range rows {
		for _, c := range row {
			switch {
			case c.Empty():
				empty++
			case c.BG == nil:
				half++
			}
		}
	}
	if empty == 0 {
		t.Error("cutout produced no fully transparent cells")
	}
	// half-transparent cells are what keep the silhouette from getting a
	// rectangular halo; if this is zero the edge handling regressed.
	if half == 0 {
		t.Error("cutout produced no half-transparent edge cells")
	}
	t.Logf("empty=%d half-transparent=%d", empty, half)
}

func TestDefaultThemeValidates(t *testing.T) {
	if err := DefaultTheme().Validate(); err != nil {
		t.Errorf("default theme must pass its own gate: %v", err)
	}
}

func TestPoseDoesNotMutateBase(t *testing.T) {
	th := DefaultTheme()
	for _, s := range Views() {
		before := make([]string, len(s.art))
		copy(before, s.art)

		poses := []Pose{s.Blink}
		for i := 0; i < s.MouthLevels(); i++ {
			poses = append(poses, s.Mouth(i))
		}
		for _, p := range poses {
			s.CellsAt(th, Quadrant, s.ColsFor(12), 12, p)
		}

		for i := range s.art {
			if s.art[i] != before[i] {
				t.Fatalf("%s: row %d of the base art changed after posing:\n  was %s\n  now %s",
					s.Name, i, before[i], s.art[i])
			}
		}
	}
}

func TestBlinkClosesTheEye(t *testing.T) {
	th := DefaultTheme()
	countPupil := func(pose Pose) int {
		grid := CellsPosed(th, Half, Width, Height, pose)
		n := 0
		for _, row := range grid {
			for _, c := range row {
				if c.FG == th.Pupil || c.BG == th.Pupil {
					n++
				}
			}
		}
		return n
	}
	open := countPupil(nil)
	shut := countPupil(SideView.Blink)
	if open == 0 {
		t.Fatal("the open-eyed bird has no pupil")
	}
	if shut >= open {
		t.Errorf("blink left %d pupil cells, open-eyed has %d; the lid is not closing",
			shut, open)
	}
}

func TestMouthTrack(t *testing.T) {
	track := MouthTrack("Hello, world.", 20)
	if len(track) == 0 {
		t.Fatal("empty track")
	}
	if track[len(track)-1] != 0 {
		t.Errorf("track ends at level %d; narration should settle shut", track[len(track)-1])
	}

	var opened, shut bool
	for _, v := range track {
		if v > 1 {
			opened = true
		}
		if v == 0 {
			shut = true
		}
		if v < 0 || v >= FrontView.MouthLevels() {
			t.Fatalf("level %d is outside the sprite's %d shapes", v, FrontView.MouthLevels())
		}
	}
	if !opened || !shut {
		t.Errorf("track never both opened and shut (opened=%v shut=%v)", opened, shut)
	}

	// deterministic: a recorded take has to be reproducible
	if a, b := MouthTrack("same line", 20), MouthTrack("same line", 20); len(a) != len(b) {
		t.Error("MouthTrack is not deterministic")
	}

	// silence is silence
	if got := MouthTrack("", 20); len(got) > 0 {
		for _, v := range got {
			if v != 0 {
				t.Error("empty text produced a moving mouth")
				break
			}
		}
	}
}

func TestFrontViewMouthsChangeTheBeak(t *testing.T) {
	th := DefaultTheme()
	// compare whole rendered grids, not a colour tally: a cell pairs two source
	// rows, so blacking one row can leave every cell still holding some red
	// while the picture has plainly changed.
	render := func(lvl int) [][]canvas.Cell {
		return FrontView.CellsAt(th, Half, FrontView.ColsFor(22), 22, FrontView.Mouth(lvl))
	}
	same := func(a, b [][]canvas.Cell) bool {
		for y := range a {
			for x := range a[y] {
				if a[y][x] != b[y][x] {
					return false
				}
			}
		}
		return true
	}

	shut := render(0)
	for lvl := 1; lvl < FrontView.MouthLevels(); lvl++ {
		if same(shut, render(lvl)) {
			t.Errorf("mouth level %d renders identically to shut", lvl)
		}
	}
	for lvl := 1; lvl < FrontView.MouthLevels()-1; lvl++ {
		if same(render(lvl), render(lvl+1)) {
			t.Errorf("mouth levels %d and %d render identically", lvl, lvl+1)
		}
	}
}

func TestGazeMovesThePupilAndSparesTheFace(t *testing.T) {
	for _, s := range Views() {
		eyes := s.eyes()
		if len(eyes) == 0 {
			t.Fatalf("%s: no eye found in the art", s.Name)
		}
		if s.Name == "front" && len(eyes) != 2 {
			t.Errorf("front view found %d eyes, want 2", len(eyes))
		}

		rest := s.Pixels(nil)
		for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {1, 1}} {
			posed := s.Pixels(s.Gaze(d[0], d[1]))

			var changed, offSocket int
			for y := range rest {
				for x := 0; x < len(rest[y]); x++ {
					if rest[y][x] == posed[y][x] {
						continue
					}
					changed++
					// anything the gaze touches must already have been eye
					if rest[y][x] != 'E' && rest[y][x] != 'X' {
						offSocket++
					}
				}
			}
			if changed == 0 {
				t.Errorf("%s: gaze %v changed nothing", s.Name, d)
			}
			if offSocket > 0 {
				t.Errorf("%s: gaze %v painted %d pixels outside the socket",
					s.Name, d, offSocket)
			}
		}

		if p := s.Gaze(0, 0); len(p) != 0 {
			t.Errorf("%s: looking straight ahead should be an empty pose", s.Name)
		}
		// out of range clamps rather than losing the pupil entirely
		far := s.Pixels(s.Gaze(9, 9))
		one := s.Pixels(s.Gaze(1, 1))
		for y := range far {
			if far[y] != one[y] {
				t.Errorf("%s: gaze is not clamped to GazeRange", s.Name)
				break
			}
		}
	}
}

func TestMouthTrackForFitsTheDuration(t *testing.T) {
	const fps = 24
	for _, secs := range []float64{0.5, 2, 10} {
		got := MouthTrackFor("hello. i am a puffin.", fps, secs)
		want := int(secs*fps + 0.5)
		if len(got) != want {
			t.Errorf("%.1fs at %dfps: got %d frames, want %d", secs, fps, len(got), want)
		}
	}
	// stretching must not invent levels the sprite cannot draw
	for _, v := range MouthTrackFor("testing one two", fps, 7) {
		if v < 0 || v >= FrontView.MouthLevels() {
			t.Fatalf("level %d outside the sprite's %d shapes", v, FrontView.MouthLevels())
		}
	}
}

func TestEmotesSettleAndEnumerate(t *testing.T) {
	for _, s := range Views() {
		names := s.Emotes()
		if len(names) == 0 {
			t.Fatalf("%s: no emotes", s.Name)
		}
		for _, n := range names {
			e, ok := s.Emote(n)
			if !ok {
				t.Fatalf("%s: Emotes() listed %q but Emote() does not have it", s.Name, n)
			}
			if e.Length() <= 0 {
				t.Errorf("%s/%s: zero length", s.Name, n)
			}
			// a non-looping emote MUST end, or a bird gets stuck in the last
			// frame of startled when the next cue arrives early
			if _, running := e.Pose(e.Length()); running != e.Loop {
				t.Errorf("%s/%s: running=%v past its end, loop=%v", s.Name, n, running, e.Loop)
			}
			if _, running := e.Pose(0); !running {
				t.Errorf("%s/%s: not running at t=0", s.Name, n)
			}
		}
		if _, ok := s.Emote("no-such-emote"); ok {
			t.Errorf("%s: unknown emote reported as present", s.Name)
		}
	}
}

func TestEmotePartsAreHonest(t *testing.T) {
	// only emotes that declare the mouth may touch it: a script driving speech
	// needs to refuse those rather than stutter the mouth track.
	for _, s := range Views() {
		for _, n := range s.Emotes() {
			e, _ := s.Emote(n)
			safe := true
			for _, f := range e.Frames {
				if f.Pose.Only(^PartMouth).Parts()&PartMouth != 0 {
					safe = false
				}
			}
			if !safe {
				t.Errorf("%s/%s: Only(^PartMouth) still yields mouth overlays", s.Name, n)
			}
			if n == "gasp" && e.Parts()&PartMouth == 0 {
				t.Errorf("%s/gasp: moves the mouth but does not declare it", s.Name)
			}
		}
	}
}

func TestNewSpriteValidates(t *testing.T) {
	good := []string{"KKWW", "KKWW"}
	if _, err := NewSprite("ok", good); err != nil {
		t.Errorf("valid sprite rejected: %v", err)
	}
	for _, c := range []struct {
		name string
		art  []string
		want string
	}{
		{"", good, "name"},
		{"empty", nil, "empty"},
		{"ragged", []string{"KKWW", "KKW"}, "rectangular"},
		{"bogus", []string{"KKZZ", "KKWW"}, "unknown role"},
	} {
		if _, err := NewSprite(c.name, c.art); err == nil {
			t.Errorf("%s: expected an error mentioning %q, got none", c.name, c.want)
		}
	}
	// an odd row count is padded, not refused: it is always the right fix
	s, err := NewSprite("odd", []string{"KK", "WW", "KK"})
	if err != nil {
		t.Fatal(err)
	}
	if _, h := s.Size(); h%2 != 0 {
		t.Errorf("odd art was not padded to an even height, got %d", h)
	}
}

func TestSpriteFileRoundTrip(t *testing.T) {
	var buf strings.Builder
	if err := WriteSprite(&buf, FrontView); err != nil {
		t.Fatal(err)
	}
	back, err := ParseSprite(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("re-reading what we wrote: %v", err)
	}
	if back.Name != FrontView.Name {
		t.Errorf("name %q, want %q", back.Name, FrontView.Name)
	}
	a, b := FrontView.Pixels(nil), back.Pixels(nil)
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("art row %d changed through a round trip", i)
		}
	}
	if len(back.Mouths) != len(FrontView.Mouths) {
		t.Errorf("mouths %d, want %d", len(back.Mouths), len(FrontView.Mouths))
	}
	// and it still renders identically
	th := DefaultTheme()
	x := FrontView.CellsAt(th, Quadrant, 14, 14, FrontView.Mouth(2))
	y := back.CellsAt(th, Quadrant, 14, 14, back.Mouth(2))
	for r := range x {
		for c := range x[r] {
			if x[r][c] != y[r][c] {
				t.Fatalf("round-tripped sprite renders differently at %d,%d", c, r)
			}
		}
	}
}

// A sprite nobody here drew must get the whole vocabulary, not just the parts
// that were wired up for the two built-ins. Reported from the puffin
// integration: NewSprite did not assemble emotes, so every cue puffin drives
// the bird with was unavailable to exactly the people the file format exists
// for.
func TestStrangerSpriteGetsTheVocabulary(t *testing.T) {
	f, err := os.Open("testdata/stranger.sprite")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	s, err := ParseSprite(f)
	if err != nil {
		t.Fatalf("parsing a hand-written sprite: %v", err)
	}
	if s.Name != "stranger" {
		t.Errorf("name %q", s.Name)
	}

	if len(s.Emotes()) == 0 {
		t.Fatal("a loaded sprite has no emotes at all")
	}
	for _, want := range []string{"blink", "startled", "curious"} {
		if _, ok := s.Emote(want); !ok {
			t.Errorf("loaded sprite is missing the %q emote", want)
		}
	}

	// gaze needs eyes found in ITS art, not in the package's
	if got := len(s.eyes()); got == 0 {
		t.Error("no eyes found in the stranger's art")
	}
	if len(s.Gaze(1, 0)) == 0 {
		t.Error("stranger cannot look sideways")
	}

	// and it renders, themed, at a size nobody chose for it
	th := DefaultTheme()
	if err := th.Validate(); err != nil {
		t.Fatal(err)
	}
	grid := s.CellsAt(th, Quadrant, s.ColsFor(10), 10, s.Gaze(1, 0))
	if len(grid) != 10 {
		t.Errorf("rendered %d rows, want 10", len(grid))
	}
}

// A sprite with no eyes must not advertise emotes it cannot perform: a cue that
// validates and then does nothing is worse than one that errors.
func TestEyelessSpriteDoesNotAdvertiseEyeEmotes(t *testing.T) {
	s, err := NewSprite("blob", []string{"KKKK", "KWWK", "KWWK", "KKKK"})
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range s.Emotes() {
		e, _ := s.Emote(n)
		empty := true
		for _, f := range e.Frames {
			// a transform-only emote has no pose and is still a performance:
			// an eyeless blob genuinely can pop and bob
			if len(f.Pose) > 0 || !f.Transform.IsIdentity() {
				empty = false
			}
		}
		if empty {
			t.Errorf("emote %q is advertised but does nothing at all", n)
		}
	}
	// it must not claim the ones it has no eyes for
	for _, n := range []string{"startled", "curious", "shifty"} {
		if _, ok := s.Emote(n); ok {
			t.Errorf("an eyeless sprite advertises %q", n)
		}
	}
}

// A sprite with an eye but no cheek around it -- Ms. Pac-Man's disc is flat
// Dark with no Light pixels anywhere near the eye -- has nowhere for the eye
// to grow into. WideEyes must come back empty rather than fall back to
// repainting the resting socket: a scrambled eye is worse than no reaction,
// the same principle TestEyelessSpriteDoesNotAdvertiseEyeEmotes holds for a
// sprite with no eye at all.
func TestCheeklessSpriteWideEyesIsInert(t *testing.T) {
	s, err := NewSprite("disc", []string{"KKKKK", "KKEKK", "KKKKK"})
	if err != nil {
		t.Fatal(err)
	}
	if p := s.WideEyes(1.6); p != nil {
		t.Errorf("wide eyes painted %d overlay(s) with no cheek to grow into", len(p))
	}
	// "surprised" is pure WideEyes and must not be advertised; "startled"
	// also touches blink and Gaze, so it can still be legitimately offered
	// with the widen beat silently dropped -- see mspacman.go for the real
	// case this covers.
	if _, ok := s.Emote("surprised"); ok {
		t.Error(`a cheekless sprite advertises "surprised", which it cannot perform`)
	}
}

func TestTransformEmotes(t *testing.T) {
	s := FrontView

	// the identity is the zero value, so every existing emote is unaffected
	for _, n := range []string{"blink", "startled", "curious"} {
		e, ok := s.Emote(n)
		if !ok {
			t.Fatalf("missing %q", n)
		}
		for i, f := range e.Frames {
			if !f.Transform.IsIdentity() {
				t.Errorf("%s frame %d gained a transform it never asked for", n, i)
			}
		}
	}

	pop, ok := s.Emote("pop")
	if !ok {
		t.Fatal("no pop emote")
	}
	maxScale, dx, dy := pop.Bounds()
	if maxScale <= 1 {
		t.Errorf("pop peaks at %.2f; it is meant to grow", maxScale)
	}
	if dx != 0 || dy != 0 {
		t.Errorf("pop translates (%d,%d); it should only scale", dx, dy)
	}

	// a caller reserving a block needs the peak BEFORE playing
	if got := ScaleRows(11, maxScale); got <= 11 {
		t.Errorf("ScaleRows(11, %.2f) = %d, want more than 11", maxScale, got)
	}
	if got := ScaleRows(1, 0.01); got != 1 {
		t.Errorf("ScaleRows floor is %d, want 1", got)
	}

	// a transform-only emote still counts as a performance
	bob, ok := s.Emote("bob")
	if !ok {
		t.Fatal("no bob emote")
	}
	if !bob.Loop {
		t.Error("bob should loop")
	}
	if _, running := bob.At(bob.Length() * 3); !running {
		t.Error("a looping emote stopped running")
	}
	if _, _, dy := bob.Bounds(); dy == 0 {
		t.Error("bob does not move vertically")
	}

	// At and Pose agree
	for _, d := range []time.Duration{0, 50 * time.Millisecond, 200 * time.Millisecond} {
		f, ok1 := pop.At(d)
		p, ok2 := pop.Pose(d)
		if ok1 != ok2 {
			t.Errorf("At and Pose disagree on running at %v", d)
		}
		if len(f.Pose) != len(p) {
			t.Errorf("At and Pose disagree on the pose at %v", d)
		}
	}
}

// Eyes wide is the strongest thing the bird has at small sizes, which is where
// it matters: at eight rows the pupil cannot travel far enough to be seen. If a
// change makes surprise quieter than a blink, that is a regression in the one
// place the corner bird depends on.
func TestWideEyesOutReadsBlink(t *testing.T) {
	th := DefaultTheme()
	// The invariant is about the views that get used SMALL -- the corner bird
	// is the profile and the presenter is the front. The turn frames exist for
	// a pan at splash size and are never asked to convey a reaction at eight
	// rows, so holding them to a corner-size legibility floor tests nothing
	// anyone depends on.
	for _, s := range []Sprite{SideView, FrontView} {
		for _, rows := range []int{8, 11, 14, 22} {
			cols := s.ColsFor(rows)
			count := func(p Pose) int {
				rest := s.CellsAt(th, Quadrant, cols, rows, nil)
				got := s.CellsAt(th, Quadrant, cols, rows, p)
				n := 0
				for y := range rest {
					for x := range rest[y] {
						if rest[y][x] != got[y][x] {
							n++
						}
					}
				}
				return n
			}
			wide, blink := count(s.WideEyes(1.6)), count(s.Blink)
			if wide <= blink {
				t.Errorf("%s at %d rows: wide eyes change %d cells, blink %d; "+
					"surprise must out-read a blink", s.Name, rows, wide, blink)
			}
		}
	}
}

func TestWideEyesStayOnTheFace(t *testing.T) {
	// this one DOES hold for every view: nothing may paint outside the face.
	for _, s := range Views() {
		rest := s.Pixels(nil)
		// a big factor must be trimmed, not spill over the cap or the beak
		posed := s.Pixels(s.WideEyes(3))
		for y := range rest {
			for x := 0; x < len(rest[y]); x++ {
				if rest[y][x] == posed[y][x] {
					continue
				}
				switch rest[y][x] {
				case RoleLight, RolePupil, RoleEyeRing, RoleStripe:
				default:
					t.Fatalf("%s: wide eyes painted over %s at %d,%d",
						s.Name, RoleName(rest[y][x]), x, y)
				}
			}
		}
		if s.WideEyes(1) != nil {
			t.Errorf("%s: factor 1 should be a no-op", s.Name)
		}
	}
}
