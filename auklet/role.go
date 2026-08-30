package auklet

// The role alphabet. This is the contract between a drawing and a theme, and
// it is the reason a sprite nobody here has seen still gets every theme for
// free: the art says which PART a pixel belongs to and never which colour it
// is, so the theme decides colour afterwards for art it has never met.
//
// The names are the puffin's. The slots are not. A gopher's buck teeth are
// RoleBeakBase, RoleBeakBand and RoleBeakTip, and its paws are RoleFeet. That
// is not a workaround; it is what the indirection buys. Three bands of
// something that sticks off the front, one colour for whatever touches the
// ground, and a dark mass, a light mass and a detail tone -- most animals are
// covered by that, and a sprite that fits the alphabet inherits every theme in
// existence, including ones written after it.
//
// If your subject genuinely needs a role that is not here, adding one means
// touching Theme, RoleColor, roleWeight and every theme at once. Do not
// silently reuse RoleStripe for something that is not a dark detail on light,
// or themes will look right on one sprite and wrong on the next.
const (
	// RoleNone is transparent: the theme's Background, or nothing at all when
	// Background is nil and the sprite is a cutout.
	RoleNone = '.'

	// RoleDark is the dark mass -- on the puffin, cap, back and collar.
	RoleDark = 'K'
	// RoleLight is the light mass -- face and belly.
	RoleLight = 'W'
	// RoleWing is a tone sitting just off RoleDark, so a large dark area does
	// not read as a flat blob. Decorative: it is the first thing to lose when
	// the sprite is shrunk.
	RoleWing = 'V'
	// RoleStripe is a small dark detail on light: brows, plates, seams.
	RoleStripe = 'D'

	// The three bands of whatever protrudes: beak, snout, muzzle, teeth.
	// RoleBeakTip is usually the largest and carries the theme's accent.
	RoleBeakBase = 'B'
	RoleBeakBand = 'Y'
	RoleBeakTip  = 'R'

	// RoleFeet is whatever touches the ground.
	RoleFeet = 'O'

	// The eye. Gaze finds these two roles in the art and moves RolePupil
	// inside the region they form, so a sprite that uses them gets looking
	// around for nothing.
	RolePupil   = 'E'
	RoleEyeRing = 'X'
)

// Roles is every byte a stencil may contain, RoleNone included.
func Roles() []byte {
	return []byte{
		RoleNone, RoleDark, RoleLight, RoleWing, RoleStripe,
		RoleBeakBase, RoleBeakBand, RoleBeakTip, RoleFeet,
		RolePupil, RoleEyeRing,
	}
}

// RoleName returns a human name for a role byte, for error messages.
func RoleName(r byte) string {
	switch r {
	case RoleNone:
		return "transparent"
	case RoleDark:
		return "dark"
	case RoleLight:
		return "light"
	case RoleWing:
		return "wing"
	case RoleStripe:
		return "stripe"
	case RoleBeakBase:
		return "beak base"
	case RoleBeakBand:
		return "beak band"
	case RoleBeakTip:
		return "beak tip"
	case RoleFeet:
		return "feet"
	case RolePupil:
		return "pupil"
	case RoleEyeRing:
		return "eye ring"
	}
	return "unknown"
}

func knownRole(r byte) bool {
	for _, k := range Roles() {
		if r == k {
			return true
		}
	}
	return false
}
