package auklet

import (
	"fmt"
	"strings"
)

// The role alphabet as a PUBLISHED ARTIFACT.
//
// Roles() and RoleName() are the authority and always have been. What did not
// exist until now is a form another repository can hold a copy of and check
// itself against, which matters because daffy does exactly that: it carries
// its own table of the eleven and a comment saying the authority lives here.
//
// A comment is not a check, and the copy had already drifted before anyone
// changed a role. daffy lists eye-ring before pupil; this package lists pupil
// before eye-ring; and daffy's own comment says its table is "in the order
// auklet lists them". Nothing broke, because the ORDER is not load-bearing
// for the wire -- a .sprite carries role characters and never names -- but a
// copy that drifts while claiming not to is the shape of the failure worth
// preventing, and it arrived by transcription rather than by intent.
//
// The direction that would actually hurt is a role's MEANING changing here.
// Then a copy exports sprites that are wrong under every theme, silently, in
// a tree whose owner did not touch anything. Publishing the alphabet does not
// stop that -- nothing across an ownership boundary can -- but it makes this
// side notice it is changing a contract, and gives the other side a fixed
// thing to test against rather than a paragraph to re-read.

// AlphabetVersion is the published form's version, separate from anything
// else here: it changes when the alphabet changes, and for no other reason.
const AlphabetVersion = 1

// Alphabet is the published form: one line per role, character then a tab
// then the name, in the order Roles() returns them, under a version header.
//
// Deliberately dull. A consumer parses it with a split, diffs it against its
// own table, and fails with a line number. Anything richer would need a
// parser, and a parser is a thing that can disagree.
func Alphabet() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# auklet role alphabet, version %d\n", AlphabetVersion)
	b.WriteString("# char\tname -- the authority is Roles() and RoleName()\n")
	for _, r := range Roles() {
		fmt.Fprintf(&b, "%c\t%s\n", r, RoleName(r))
	}
	return b.String()
}
