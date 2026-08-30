# puffin-auklet

A puffin is an easy bird to recognize and a hard one to draw small. By the time
you are down to a terminal cell grid there is no room for a bird at all -- only
room for the four things that make it that bird: a black cap, a white face, a
huge banded wedge of a beak, and orange feet. Everything else is texture, and
texture is the first thing the grid takes away.

So this is a sprite that keeps those four, and a workbench for looking at it.

The art here is a stencil. It stores which *part* of the bird each pixel belongs
to -- cap, face, beak tip -- and never which colour that part is. A theme
supplies the ink. Swap the theme and the same stencil prints a corvid puffin, a
vaporwave puffin, or a bladerunner puffin, and nothing about the art has to know
that happened.

## Run it

    go run ./cmd/auklet

| key | does |
|---|---|
| `f` | cycle focus: auklet, window, view |
| arrows, `hjkl` | move whatever has focus |
| shift + arrows | move it by 5 |
| `+` / `-` | sprite size |
| `g` | glyph set |
| `[` `]` `{` `}` | window width, window height |
| `tab` | next theme |
| `c` | toggle cutout |
| `b` | cycle backdrop |
| `v` | toggle the validator pane |
| `r` | reset |
| `q`, `esc`, `ctrl-c` | quit |

It opens in cutout mode over the text backdrop, so the transparency is visible
in the first frame rather than something you have to go find.

## World, window, screen

Three coordinate spaces, and keeping them apart is most of the design.

| space | what it is |
|---|---|
| world | where the bird actually is. Unbounded. Never clamped. |
| window | a rectangle of world space that is currently being shown |
| screen | where that window is drawn in the terminal |

The bird's world position survives being scrolled out of view, having the window
moved off it, or being walked off the edge itself. It is not clipped to the
window, hidden, or snapped back -- it is somewhere the window is not looking, and
the border grows a `<`, `>`, `^` or `v` pointing at it. Move the window onto
those coordinates and the whole bird is still there.

The window is kept on screen. The bird is not. That asymmetry is deliberate: a
window you can lose is a bug, a sprite you can lose is a feature.

## Using the sprite

    t := themes.Dodo["vaporwave"].Puffin()
    t.Background = nil            // nil background means cutout

    c := canvas.New(w, h)
    c.Fill(canvas.Cell{R: ' ', BG: pageBg})
    c.Blit(puffin.CellsAt(t, puffin.Quadrant, puffin.ColsFor(12), 12), x, y)
    fmt.Print(c.String())

`puffin.Render(t)` returns the bird as a plain string at scale 1 if you do not
need to composite.

Compositing MUST go through `canvas`. Styled strings cannot be overlaid by
concatenation: the escape sequences carry no geometry, so there is nothing to
clip a sprite against. The buffer composes in cell space and produces the string
once, at the end.

## Roles

| role | part of the bird |
|---|---|
| `Background` | the field; nil makes the sprite a cutout |
| `Dark` | crown, back, throat collar |
| `Light` | face patch and belly |
| `Wing` | tone on the folded wing, sits just off `Dark` |
| `BeakBase` | the slate section against the face |
| `BeakBand` | the ridge stripe and the gape line |
| `BeakTip` | the outer half |
| `Feet` | feet |
| `Pupil`, `EyeRing` | eye |
| `Stripe` | eye plate and the line trailing to the nape |

## Themes come from dodo

The four themes are dodo's own, read out of dodo's stylesheet. No hex is copied
into Go by hand.

    go run ./cmd/gen-themes ../../dodo/lib/themes/themes.css > themes/dodo_gen.go

`themes/dodo_gen.go` is generated. Do not hand-edit it. The generator fails
loudly if a variable the adapter needs has gone missing, rather than emitting a
theme with holes in it.

The adapter is not a rename. Dodo names its colours for their job in a *page* --
`--ink` is body text -- so on a dark theme `--ink` is nearly white and on stardew
it is nearly black. The bird needs luminance roles, not page roles, so the
adapter branches on how bright `--bg` is and picks accordingly. Map `--ink` onto
`Light` naively and stardew renders a dark bird with a darker face.

Semantic colours (`--ok`, `--warn`, `--caution`) are deliberately unused. A
warning colour spent on a beak stops reading as a warning.

## Validation

The bird is legible because of contrast *relationships*, not because of any
particular hue. An arbitrary theme breaks those relationships silently -- it
still renders, it just renders a blob. `Theme.Validate()` reports every collapsed
pair at once:

    Light is not bright enough against Dark (luminance gap 0.200, need 0.300):
        the face and belly stop separating from the cap
    BeakTip and BeakBand are too close (0.000, need 0.150):
        the beak stops reading as banded, which is the puffin's whole signature

Call it once when the theme loads, not per frame.

`Light` against `Dark` is gated on luminance; a mid-grey and a mid-blue sit far
apart in RGB and still read as one flat mass at cell size. Every other pair is
gated on distance.

Validation reads the hex you declared, not `RGBA()`. `RGBA()` resolves through
the renderer's colour profile and returns black when there is no TTY, which would
make every theme identical and passing.

### What it cannot see

Two honest gaps, both of which fail open:

- A palette-index colour (`lipgloss.Color("205")`) carries no readable RGB. Those
  pairs are reported as **unverifiable** rather than passed, but a 16-colour
  theme gets no real gate. Check it by eye.
- Validation checks declared colours, not what survives degradation to a smaller
  palette. Two distinct near-blacks can both flatten to colour 0 on an ANSI16
  terminal and the bird disappears with a clean bill of health.

## Size

`CellsAt(theme, glyphs, cols, rows)` renders to any size. `ColsFor(rows)` gives
the column count that keeps the bird's proportions -- a terminal cell is about
twice as tall as it is wide, so a picture that is square on screen is not square
in cells, and that ratio is applied in exactly one place.

Two things make shrinking work rather than turning the bird to mush.

**More pixels per cell.** A cell can hold more than the two stacked pixels of a
half block. Every set below is complete -- there is a glyph for every possible
arrangement of lit subcells -- so the *shape* is always exact and only the colour
is approximated, down to the two a cell can carry.

| set | subcells | notes |
|---|---|---|
| half | 1x2 | universal |
| quadrant | 2x2 | Unicode 1.1; universal in practice. The default. |
| sextant | 2x3 | Unicode 13 (2020). Best small, needs a font that has them. |

**A weighted vote.** When one subcell has to stand for several source pixels, a
plain majority is the wrong rule: the eye gets outvoted by the face around it and
the beak's ridge stripe gets outvoted by the beak, and what survives is a blob
with a wedge on it. So roles carry weights, and the ones that carry the bird's
identity shout louder than the ones that are merely large. That is the same
judgement the art encodes -- keep the cap, the face, the banded beak, the feet --
applied to sampling. The weights are in `puffin/resample.go` and they are the
first thing to reach for if a feature disappears too early.

It stays legible down to about eight rows on sextants. Below that it is a dark
bird with an orange wedge, which is still more puffin than most things.

## Constraints worth knowing before you change something

A cell holds two colours. Not three. Every part of the pipeline that looks
clever is working around that one number.

A cell holds two pixels stacked, drawn with a half-block. Which glyph a cell gets
depends on which halves survive -- both opaque is the ordinary
foreground-over-background pair, only the top gives `▀` with no background, only
the bottom gives `▄`. That is what lets a cutout keep a clean silhouette instead
of a rectangular halo, and it is the part to be careful with when editing
`puffin.Cells`.

## Layout

| path | what |
|---|---|
| `puffin/` | the stencil, the theme, the validator |
| `canvas/` | cell buffer and compositing |
| `themes/` | dodo's four plus three diagnostic themes |
| `scene/` | backdrop and placement, shared by the viewer and the renderers |
| `cmd/auklet` | the viewer |
| `cmd/gen-themes` | stylesheet to Go |
| `cmd/shot` | renders one composed screen to stdout, no TTY needed |
| `cmd/dump` | one line per cell, for checking a render without parsing escapes |
| `cmd/sizes` | dumps the sprite alone at a given glyph set and row count |

Three of the seven themes exist to fail. `nord` is a real palette that misses the
silhouette gate by three thousandths, `ansi16` demonstrates the unverifiable
path, and `collapsed` trips eight gates at once. A picker that only ever shows
passing themes teaches you nothing about the gate.
