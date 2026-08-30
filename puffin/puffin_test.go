package puffin

import "testing"

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
	before := make([]string, len(art))
	copy(before, art)

	th := DefaultTheme()
	CellsPosed(th, Quadrant, ColsFor(12), 12, Pose{Blink})

	for i := range art {
		if art[i] != before[i] {
			t.Fatalf("row %d of the base art changed after posing:\n  was %s\n  now %s",
				i, before[i], art[i])
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
	shut := countPupil(Pose{Blink})
	if open == 0 {
		t.Fatal("the open-eyed bird has no pupil")
	}
	if shut >= open {
		t.Errorf("blink left %d pupil cells, open-eyed has %d; the lid is not closing",
			shut, open)
	}
}
