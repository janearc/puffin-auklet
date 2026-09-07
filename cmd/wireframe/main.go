// wireframe fills a daffy layout with auklet backdrops.
//
// daffy draws and this renders. That separation is the point: daffy will not
// run code for you, and it should not -- a drawing tool that executes what is
// in a document is a drawing tool you cannot open a stranger's file with. So
// the document says WHERE, by name, and this says WITH WHAT.
//
// The seam it uses was already there. daffy's `export json` writes the regions
// and the theme as data -- "for whatever consumes them next", in its own words
// -- and scene.Backdrop takes a canvas rather than a position, so it has never
// known where on screen it is. Carve a child canvas per region, fill it, blit
// it back. Canvas.Blit already clips at every edge and skips empty cells, so
// transparency and overlap come for free.
//
// Usage:
//
//	wireframe -layout roster.json                  # fills from each region's notes
//	wireframe -layout roster.json -fill outer=lights -fill side=fireplace
//	wireframe -layout roster.json -frame 7         # a later phase of the twinkle
//	wireframe -layout roster.json -list            # what the file contains
//
// A region is filled if -fill names it, else if its notes carry `fill`. A
// region with neither is left alone rather than guessed at, because an
// unfilled region and a region filled with the wrong thing look identical on
// a first run and only one of them is honest.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/janearc/puffin-auklet/auklet"
	"github.com/janearc/puffin-auklet/canvas"
	"github.com/janearc/puffin-auklet/scene"
	"github.com/janearc/puffin-auklet/themes"
)

// Layout is daffy's `export json`, in the fields this needs.
//
// Declared here rather than imported so that this tool does not depend on
// daffy's package: the JSON is the published contract between them, and a
// shared struct would make two repositories move together for no gain.
// Unknown fields are ignored, which is what lets daffy add to it without
// breaking this.
type Layout struct {
	Name    string   `json:"name"`
	Cols    int      `json:"cols"`
	Rows    int      `json:"rows"`
	Theme   string   `json:"theme"`
	Regions []Region `json:"regions"`
}

// Region is one named area. Cells is [x0, y0, x1, y1] in daffy's order.
type Region struct {
	Name  string            `json:"name"`
	Kind  string            `json:"kind"`
	Cells [4]int            `json:"cells"`
	Shape bool              `json:"shape"`
	Notes map[string]string `json:"notes"`
}

// Size is the region's width and height in cells, never negative.
func (r Region) Size() (w, h int) {
	w, h = r.Cells[2]-r.Cells[0], r.Cells[3]-r.Cells[1]
	return max(w, 0), max(h, 0)
}

// fills collects repeated -fill name=backdrop flags.
type fills map[string]string

func (f fills) String() string { return "" }

func (f fills) Set(v string) error {
	name, back, ok := strings.Cut(v, "=")
	if !ok {
		return fmt.Errorf("want name=backdrop, got %q", v)
	}
	f[name] = back
	return nil
}

func main() {
	path := flag.String("layout", "", "a daffy `export json` file")
	frame := flag.Int("frame", 0, "the animation frame to draw; backdrops phase on it")
	theme := flag.String("theme", "", "auklet theme to draw with; the layout's own name is tried first")
	list := flag.Bool("list", false, "describe the layout and exit")
	fill := fills{}
	flag.Var(fill, "fill", "region=backdrop, repeatable")
	flag.Parse()

	if *path == "" {
		fmt.Fprintln(os.Stderr, "wireframe: -layout is required")
		fmt.Fprintln(os.Stderr, "backdrops:", strings.Join(scene.BackdropNames, " "))
		os.Exit(2)
	}
	l, err := load(*path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "wireframe:", err)
		os.Exit(1)
	}
	if *list {
		describe(os.Stdout, l, fill)
		return
	}
	t, used := pickTheme(l.Theme, *theme)
	out, warns := render(l, fill, t, *frame)
	fmt.Print(out)
	// the theme actually used is named on stderr, because a layout asking for
	// a theme this tree does not have is a real difference between the picture
	// daffy showed and the one here, and silence would hide it
	fmt.Fprintf(os.Stderr, "%s  %dx%d  theme %s  frame %d\n", l.Name, l.Cols, l.Rows, used, *frame)
	for _, w := range warns {
		fmt.Fprintln(os.Stderr, "  "+w)
	}
}

// load reads and validates a layout.
func load(path string) (*Layout, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var l Layout
	if err := json.Unmarshal(body, &l); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if l.Cols <= 0 || l.Rows <= 0 {
		return nil, fmt.Errorf("%s: layout is %dx%d", path, l.Cols, l.Rows)
	}
	return &l, nil
}

// pickTheme honours the layout's theme name when this tree has one by that
// name, and says which it settled on either way.
func pickTheme(want, override string) (auklet.Theme, string) {
	for _, name := range []string{override, want} {
		if name == "" {
			continue
		}
		for _, n := range themes.All {
			if strings.EqualFold(n.Name, name) {
				return n.Theme, n.Name
			}
		}
	}
	return themes.All[0].Theme, themes.All[0].Name + " (fallback)"
}

// render draws every region that has a fill, and reports what it could not do
// rather than dropping it.
func render(l *Layout, fill fills, t auklet.Theme, frame int) (string, []string) {
	var warns []string
	root := canvas.New(l.Cols, l.Rows)

	for _, r := range l.Regions {
		name := fillFor(r, fill)
		if name == "" {
			continue
		}
		w, h := r.Size()
		if w == 0 || h == 0 {
			warns = append(warns, fmt.Sprintf("region %q is %dx%d, nothing drawn", r.Name, w, h))
			continue
		}
		// a child canvas the size of the region. The backdrop never learns
		// where this sits, which is why the same function works as a
		// full-screen field and as a 16x16 corner.
		child := canvas.New(w, h)
		if !scene.Backdrop(child, name, t, nil, frame) {
			warns = append(warns, fmt.Sprintf("region %q wants backdrop %q, which is not one of: %s",
				r.Name, name, strings.Join(scene.BackdropNames, " ")))
			continue
		}
		root.BlitCanvas(child, r.Cells[0], r.Cells[1])
		if r.Shape {
			// daffy says this region has a pixel mask and the JSON does not
			// carry it. Drawn as its bounding box, and said out loud, because
			// a rectangle where a shape was meant is exactly the kind of wrong
			// that looks fine.
			warns = append(warns, fmt.Sprintf("region %q has a mask daffy did not export; drawn as its box", r.Name))
		}
	}
	return root.String() + "\n", warns
}

// fillFor is the backdrop a region should get: the flag first, then the
// region's own `fill` note, then nothing.
func fillFor(r Region, fill fills) string {
	if v, ok := fill[r.Name]; ok {
		return v
	}
	return r.Notes["fill"]
}

// describe prints what the file contains without drawing it, so a wireframe
// can be inspected before it is filled.
func describe(w *os.File, l *Layout, fill fills) {
	fmt.Fprintf(w, "%s  %dx%d  theme %s  %d region(s)\n", l.Name, l.Cols, l.Rows, l.Theme, len(l.Regions))
	rs := append([]Region(nil), l.Regions...)
	sort.Slice(rs, func(i, j int) bool { return rs[i].Name < rs[j].Name })
	for _, r := range rs {
		cw, ch := r.Size()
		f := fillFor(r, fill)
		if f == "" {
			f = "-"
		}
		shape := ""
		if r.Shape {
			shape = "  (masked)"
		}
		fmt.Fprintf(w, "  %-16s %-10s %3d,%-3d %3dx%-3d  fill %s%s\n",
			r.Name, r.Kind, r.Cells[0], r.Cells[1], cw, ch, f, shape)
	}
	fmt.Fprintln(w, "backdrops:", strings.Join(scene.BackdropNames, " "))
}
