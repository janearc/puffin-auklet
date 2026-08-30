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
// The table above is correct for DARK themes and will bite you on a light one.
// Reported from the integration, on stardew: puffin darkens its Snow token on a
// pale ground so the bird stays visible against the page, which is right -- and
// that collapses Snow against Plumage, so Light and Dark end up 0.041 apart
// against a floor of 0.300. Two correct adjustments colliding.
//
// So prefer themes.Dodo[name].Auklet() where the theme is one of dodo's four:
// it branches on ground brightness and picks different fields for a light
// ground. Fall back to the hand mapping only for themes that adapter has not
// seen, and let Validate tell you whether it worked.
//
// puffin's Theme.Light already knows that a pale ground needs different
// choices, for the same reason this package cares. Trust that flag rather than
// re-deriving it. If you ever do need to measure, Luminance is exported.
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
// narration. It is deterministic, so a take reproduces exactly. Against real
// synthesised audio use MouthTrackFor, which fits the track to a known duration
// -- MouthTrack derives its length from the text, which is only a guess at how
// long the line takes to say, and a mouth that stops two seconds before the
// voice does is worse than one that never moved.
//
// Gaze shifts the pupils, so the bird can look at whatever the application is
// attending to. scene.GazeAt turns a world-space target into that pose.
//
// # What can and cannot move
//
// A part can move exactly as far as there is art behind it.
//
// The pupil can move: the socket it sits in is already drawn, so shifting it
// uncovers ring rather than a hole. That is Gaze, and it is why looking around
// cost forty lines.
//
// The beak cannot. The base art is one flat grid with nothing drawn behind
// anything else, so sliding it leaves a hole in the head. That needs layers
// with z-order and roughly 150 pixels of occluded head nobody has drawn, and it
// is an art job before it is a code one.
//
// Between those two, overlays cover what a mascot in a status bar actually
// does: blink, talk, and look at things.
//
// # Making your own bird
//
// NewSprite is the way in for art that did not ship here. It validates rather
// than trusts: a ragged row or a stray character fails there, loudly, instead
// of rendering a hole three sizes later.
//
// ParseSprite reads a sprite from a text file, which matters more than it
// looks. The point of publishing this is that somebody who does not write Go
// can draw a bird, put it in a gist, and have it work.
//
// The role alphabet in role.go is the contract that makes a stranger's bird
// themeable. The names are the puffin's; the slots are not. A gopher's buck
// teeth are the three beak roles and its paws are RoleFeet.
//
// # Emotes
//
// An Emote is a named frame series: poses with per-frame holds. Sprite.Emotes
// enumerates them, so a script can be validated before a recording rather than
// during it.
//
// They settle by construction. Emote.Pose returns (nil, false) once a
// non-looping series is over, so a bird cannot be stuck in the last frame of
// "startled" because the next cue arrived early. Interruption is likewise
// nothing special: stop asking for poses and the bird is at rest on the next
// frame, because a pose is an overlay rather than a state.
//
// Overlays declare which Part they drive, so speech and emotes can be checked
// for collision instead of silently fighting. The rule: THE MOUTH TRACK OWNS
// THE MOUTH WHILE SPEECH IS PLAYING. An emote during narration should be
// filtered:
//
//	pose := emote.Pose(elapsed).Only(^auklet.PartMouth)
//
// Emote.Parts says up front whether a series would collide, so a caller can
// refuse a cue rather than stutter the track.
//
// A Frame may also carry a Transform: a translation in cells and a size
// multiplier, identity by default. That is how a settle, a walk-on bob and the
// anime squash-and-stretch pop are one mechanism rather than three hand-rolled
// ones -- none of them is art, since the sprite already resamples to any size.
//
// The sprite does not apply a transform; the CALLER does, because the caller
// owns position and size. Scale is relative for the same reason: an emote that
// demands 13 rows fights a corner that has 8, while 1.2x composes with it. Use
// ScaleRows, and reserve the block at Emote.Bounds() before playing, or a
// sprite that grows mid-series tears whatever it is spliced into.
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
