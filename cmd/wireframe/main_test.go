package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janearc/puffin-auklet/themes"
)

// The seam between daffy and this tree, tested at the seam: a layout is a
// JSON file somebody else's program wrote, so the cases that matter are the
// ones where that file is not what this reader expects.

// writeLayout puts a layout on disk and returns its path.
func writeLayout(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "layout.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestVersionNote covers the three answers: known, absent, and newer than
// this reader.
func TestVersionNote(t *testing.T) {
	if note := (&Layout{Daffy: LayoutVersion}).VersionNote(); note != "" {
		t.Fatalf("the version this reader was written against said something: %q", note)
	}
	if note := (&Layout{}).VersionNote(); !strings.Contains(note, "no version") {
		t.Fatalf("an unversioned export should say so, got %q", note)
	}
	note := (&Layout{Daffy: LayoutVersion + 1}).VersionNote()
	if !strings.Contains(note, "check it against daffy") {
		t.Fatalf("a newer export should point at daffy, got %q", note)
	}
}

// TestLoadRendersAnUnknownVersion is the decision itself, not the message:
// a version from the future must still produce a layout. Refusing here would
// break this harness on a daffy release that only added a field.
func TestLoadRendersAnUnknownVersion(t *testing.T) {
	p := writeLayout(t, `{"daffy":99,"name":"future","cols":10,"rows":4,"regions":[]}`)
	l, err := load(p)
	if err != nil {
		t.Fatalf("a future version was refused: %v", err)
	}
	if l.Cols != 10 || l.Name != "future" {
		t.Fatalf("the version-1 fields did not survive: %+v", l)
	}
	if l.VersionNote() == "" {
		t.Fatal("rendered a future version silently, which is the one thing that must not happen")
	}
}

// TestLoadRefusesNonsenseGeometry: a version is advisory, a size is not.
// Nothing can be drawn in zero columns and the message should say the size.
func TestLoadRefusesNonsenseGeometry(t *testing.T) {
	p := writeLayout(t, `{"daffy":1,"name":"flat","cols":0,"rows":4,"regions":[]}`)
	_, err := load(p)
	if err == nil {
		t.Fatal("a zero-column layout loaded")
	}
	if !strings.Contains(err.Error(), "0x4") {
		t.Fatalf("the error should carry the size, got %v", err)
	}
}

// TestFillForPrefersTheFlag: -fill overrides the file, because the flag is
// what a person just typed and the note is what they typed last week.
func TestFillForPrefersTheFlag(t *testing.T) {
	r := Region{Name: "berth", Notes: map[string]string{"fill": "lights"}}
	if got := fillFor(r, fills{}); got != "lights" {
		t.Fatalf("the note was not read: %q", got)
	}
	if got := fillFor(r, fills{"berth": "gopher"}); got != "gopher" {
		t.Fatalf("the flag did not win: %q", got)
	}
	if got := fillFor(Region{Name: "bare"}, fills{}); got != "" {
		t.Fatalf("an unfilled region asked for %q", got)
	}
}

// TestPickThemeFallsBackAndSaysSo. A layout naming a theme this tree does not
// have is a real difference between daffy's picture and this one, so the name
// returned has to admit the substitution rather than report the theme asked
// for.
func TestPickTheme(t *testing.T) {
	first := themes.All[0].Name
	if _, used := pickTheme("no-such-theme", ""); used != first+" (fallback)" {
		t.Fatalf("an unknown theme reported %q", used)
	}
	if _, used := pickTheme(strings.ToUpper(first), ""); used != first {
		t.Fatalf("theme names should match without regard to case, got %q", used)
	}
	// the override beats the layout: it is the later instruction
	if len(themes.All) > 1 {
		second := themes.All[1].Name
		if _, used := pickTheme(first, second); used != second {
			t.Fatalf("-theme did not override the layout, got %q", used)
		}
	}
}

// TestBuildWarnsRatherThanRefuses. Every bad region here is one found in a
// real file, and the shared rule is that the picture is still drawn: a
// wireframe that refuses to render tells you less than one that renders
// wrong and says which part is wrong.
func TestBuildWarns(t *testing.T) {
	l := &Layout{Cols: 40, Rows: 12, Regions: []Region{
		{Name: "collapsed", Cells: [4]int{0, 0, 1, 1}, Notes: map[string]string{"fill": "lights"}},
		{Name: "empty", Cells: [4]int{5, 5, 5, 5}, Notes: map[string]string{"fill": "lights"}},
		{Name: "unknown", Cells: [4]int{0, 2, 20, 8}, Notes: map[string]string{"fill": "not-a-backdrop"}},
		{Name: "masked", Cells: [4]int{20, 2, 38, 10}, Shape: true, Notes: map[string]string{"fill": "lights"}},
		{Name: "quiet", Cells: [4]int{0, 9, 10, 11}},
	}}
	c, warns := build(l, fills{}, themes.All[0].Theme, 0)
	if w, h := c.Size(); w != 40 || h != 12 {
		t.Fatalf("canvas is %dx%d, want the layout's size", w, h)
	}
	joined := strings.Join(warns, "\n")
	for _, want := range []string{"collapsed drag", "nothing drawn", "neither a backdrop", "mask daffy did not export"} {
		if !strings.Contains(joined, want) {
			t.Errorf("no warning about %q; got:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, `"quiet"`) {
		t.Errorf("a region asking for no fill was warned about:\n%s", joined)
	}
}

// TestRegionSizeIsNeverNegative. daffy normalises a drag, but this reads a
// file rather than a drag, and a negative size would index backwards through
// a canvas.
func TestRegionSizeIsNeverNegative(t *testing.T) {
	w, h := Region{Cells: [4]int{9, 9, 2, 3}}.Size()
	if w != 0 || h != 0 {
		t.Fatalf("an inverted region measured %dx%d", w, h)
	}
}
