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
