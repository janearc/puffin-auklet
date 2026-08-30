package auklet

import (
	"testing"

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
