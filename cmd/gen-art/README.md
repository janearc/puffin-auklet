# Adding a sprite

The art is generated, not hand-typed. Placing characters by hand means
recounting forty rows every time a proportion changes; describing the subject as
overlapping geometry means changing one number and looking again.

    go run ./cmd/gen-art

This rewrites `auklet/*_art.go`. Those files are generated — do not hand-edit
them. Run it and diff: the grids either match what is committed or you have
found a regression.

## The loop

You cannot judge this by reading the code, and neither can whoever writes it.
Render and look, every time.

1. **Say what is distinctive first, in prose.** Not what the subject looks like
   — what *survives* being shrunk to twenty rows. Three to five features, named
   anatomically. "Big teeth" is not a specification; "two incisors, front-facing,
   about a fifth of the head's height, pale against a dark muzzle" is.
   Everything else is texture, and texture is the first thing the grid takes.
2. **Write a `draw<Name>() *drawing` in a new file here.** Fill the silhouette
   solid first, then carve the light regions *out* of it. Painting light on dark
   instead loses every boundary that is defined by dark surviving between two
   light areas — on the puffin that is the throat collar, and it is half the read.
3. **Render and look.** `go run ./cmd/sizes` dumps cells; the PNG path in
   `cmd/render` is quicker for eyeballing shape. Fix one thing, render again.
   Three to six passes is normal.
4. **Silhouette test.** Fill the whole shape flat. If it is not recognisable in
   one colour, the proportions are wrong and colour will not save it.
5. **Check it small.** 22 rows is easy. Look at 11 and 8. If a feature vanishes,
   raise its weight in `auklet/resample.go` rather than enlarging it.

## Roles

The role bytes are the contract between art and theme. Reuse them:

| byte | role | what it is |
|---|---|---|
| `.` | — | transparent |
| `K` | Dark | the dark mass: cap, back, collar |
| `W` | Light | the light mass: face, belly |
| `V` | Wing | a tone just off `K`, so the dark is not a flat blob |
| `D` | Stripe | small dark detail on light: brows, plates, seams |
| `B` | BeakBase | first band of the feature that sticks out |
| `Y` | BeakBand | second band, usually thin |
| `R` | BeakTip | third band, usually the largest |
| `O` | Feet | feet |
| `E` | Pupil | |
| `X` | EyeRing | |

The names are the puffin's; the **slots** are not. A gopher's buck teeth are
`B`/`Y`/`R` and its paws are `O`. That is not a hack — it is what "the art never
names a colour" buys you, and it means a new sprite inherits every theme,
including dodo's four, for free.

If a subject genuinely needs a role that does not exist, add it to `Theme`,
`colorFor`, `roleWeight` and every theme at once. Do not quietly reuse `D` for
something that is not a dark detail on light, or themes will look correct on one
sprite and wrong on the next.

## Registering it

Add the draw function to the `jobs` table in `main.go`, then a `Sprite` in
`auklet/sprite.go`:

    var GopherView = Sprite{Name: "gopher", art: gopherArt, Blink: gopherBlink}

and add it to `Views()`. Poses are per-sprite and their coordinates are relative
to the **trimmed** grid — the generator prints the trim offsets, and changing the
art can move them. That is why the tests compare rendered output rather than
eyeballing a map.

## Weights

`roleWeight` in `auklet/resample.go` decides what survives shrinking. When one
subcell has to stand for several source pixels, a plain majority loses the eye to
the face around it. If your subject's signature feature is small — teeth, an eye,
a stripe — give it a high weight. It is the first knob to reach for and it is
cheaper than redrawing.
