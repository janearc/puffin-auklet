// Package auklet draws an atlantic puffin in terminal cells.
//
// Notes for whoever is wiring this into the puffin TUI. Read the operational
// note first; the rest is background you can take as needed.
//
// # Getting the module
//
// github.com/janearc/puffin-auklet is a private repository. Credentials are not
// the problem -- it is the same account -- but the module proxy and the
// checksum database cannot see it, so a plain "go get" fails on a fetch rather
// than on a login:
//
//	GOPRIVATE=github.com/janearc/*
//
// Simpler again, while both trees sit on the same disk: a replace directive in
// puffin's go.mod, which skips the proxy entirely.
//
//	require github.com/janearc/puffin-auklet v0.0.0
//
//	replace github.com/janearc/puffin-auklet => ../tools/puffin-auklet
//
// Adjust the relative path to wherever puffin sits. Both live under ~/mesh/dev,
// and ~/mesh/whoops is a symlink to the same tree, so either spelling works.
// The replace also means a change here is visible to puffin immediately, with
// no tag-and-fetch round trip, which is what you want while this is moving.
//
// Work lands on auklet-dev. main carries only the initial commit.
//
// # Why the package is not called puffin
//
// Because the repository you are working in is, and it already has a Theme.
// Two types called Theme in one file is a bad afternoon. The repo is
// puffin-auklet; the package is auklet; the workbench binary is bandersnatch.
//
// # Mapping puffin's theme onto the bird
//
// The adapter has to live on puffin's side, since nothing can import a package
// main. It is close to a rename, because puffin's Theme already carries the
// bird's own colours:
//
//	auklet role                  puffin theme field
//	Background                   Bg, or nil for a cutout
//	Dark                         Plumage
//	Light                        Snow
//	Wing                         Line
//	BeakBase                     Dim
//	BeakBand                     AccentToken
//	BeakTip, Feet, EyeRing       Accent
//	Pupil                        Plumage
//	Stripe                       Dim
//
// Do not map Ok, Warn or Caution onto the bird. They are semantic, they are
// meant to read the same in every skin, and a warning colour spent on a beak
// stops being a warning.
//
// puffin's Theme.Light already knows that a pale ground needs different
// choices, and for the same reason this package cares: the bird's white face
// vanishes on one. Trust that flag rather than re-deriving it. If you ever do
// need to measure, Luminance is exported.
//
// Call Theme.Validate once when the theme is selected. Not per frame. It
// reports every collapsed contrast pair at once, with a sentence on what each
// one costs, because a broken theme still renders -- it just renders a blob.
//
// # Compositing
//
// Do not overlay sprites by concatenating styled strings. Escape sequences
// carry no geometry, so there is nothing to clip against and the result is
// wrong in a way that looks like a font problem. Use the canvas package: it is
// a cell buffer, Blit composites cells into cells with transparency, and the
// string is produced once at the end.
//
//	c := canvas.New(w, h)
//	c.Fill(canvas.Cell{R: ' ', BG: pageBg})
//	c.Blit(auklet.FrontView.CellsAt(t, auklet.Quadrant, cols, rows, pose), x, y)
//	fmt.Print(c.String())
//
// A nil Theme.Background makes the sprite a cutout: cells outside the bird are
// skipped entirely, and half-cells along the silhouette keep whatever is behind
// them, so there is no rectangular halo.
//
// Rendering is cheaper than it looks. canvas.Render run-length-encodes adjacent
// cells that share a style, so a frame is a few hundred bytes, and bubbletea
// only writes what changed. An idle bird that blinks every few seconds costs
// almost nothing.
//
// # Sizing
//
// CellsAt renders to any size. ColsFor gives the column count that preserves
// the bird's proportions -- a terminal cell is about twice as tall as it is
// wide, and that ratio is applied in exactly one place, so do not apply it
// again yourself.
//
// Shrinking works, down to about eight rows. Two things make that true: richer
// glyph sets carry more pixels per cell (Quadrant is the default and is
// universally supported; Sextant is better small but needs a font with
// U+1FB00), and the resampler uses a weighted vote so small load-bearing
// features are not averaged away. Those weights are in resample.go and they are
// the first thing to reach for if a feature disappears too early.
//
// Dithering was tried and reverted. It is worse: at this size there is no
// spatial resolution left to trade for colour resolution, and it undoes the
// weighting on purpose-built features.
//
// # Views and poses
//
// SideView is the profile and reads at the smallest sizes, because the beak is
// doing all the work. FrontView faces the reader and is the one that can talk;
// head-on the beak is a narrow blade rather than a wedge, since it is flattened
// side to side.
//
// Poses belong to a sprite, not to the package, because they are coordinates
// rather than concepts. Blink in profile is one eyelid in one place; head-on it
// is two, somewhere else. Using an overlay from the wrong view puts a lid on a
// cheek.
//
// MouthTrack turns a line of text into one mouth level per frame, for
// narration. It is deterministic, so a take reproduces exactly.
//
// What is NOT supported is moving a part. The base art is one flat grid with
// nothing drawn behind anything else, so sliding the beak leaves a hole in the
// head. Real articulation needs layers with z-order and roughly 150 pixels of
// occluded head that nobody has drawn. Overlays cover blinking and talking,
// which is what a mascot in a status bar actually does.
//
// # Relationship to splash.go
//
// puffin's existing splash already separates beak from plumage so the beak can
// wear the orange. This is the same instinct generalised: eleven roles instead
// of two, and the art naming no colours at all. Replacing puffinRows with a
// sprite gets theming, arbitrary size, cutout, blink and the front view. Do not
// try to feed the old rows into this package -- they are a different encoding.
//
// # Known gaps, both of which fail open
//
// Validate reads the colours a theme declares, not what survives degradation to
// a smaller palette. Two distinct near-blacks can both flatten to colour 0 on a
// 16-colour terminal and the bird disappears with a clean bill of health.
//
// A palette-index colour carries no readable RGB, so those pairs are reported
// as unverifiable rather than passed. A 16-colour theme therefore gets no real
// gate; check it by eye.
//
// # Looking at your changes
//
// go run ./cmd/bandersnatch is the workbench: every theme, arbitrary size, the
// glyph sets, cutout against backdrops that make transparency obvious, a window
// to move the bird in and out of, narration and the intro. cmd/shot and
// cmd/dump render without a TTY if you need to check something in a test.
//
// themes/dodo_gen.go is generated from dodo's stylesheet by cmd/gen-themes. Do
// not hand-edit it.
package auklet
