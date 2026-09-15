package auklet

import (
	"os"
	"strings"
	"testing"
)

// The published alphabet is exactly what the code says, and stays that way.
//
// This is the check that gives the publication its point: if a role's name or
// character changes, or one is added or removed, this test fails HERE, in the
// tree that owns the alphabet, rather than in a consumer that copied it. That
// is the whole mechanism -- it cannot force a copy to update, and nothing
// across an ownership boundary can, but it stops this side changing a
// published contract without noticing.
func TestAlphabetMatchesTheCode(t *testing.T) {
	lines := strings.Split(strings.TrimSuffix(Alphabet(), "\n"), "\n")
	var got []string
	for _, l := range lines {
		if !strings.HasPrefix(l, "#") {
			got = append(got, l)
		}
	}
	roles := Roles()
	if len(got) != len(roles) {
		t.Fatalf("alphabet lists %d roles, Roles() returns %d", len(got), len(roles))
	}
	for i, r := range roles {
		char, name, ok := strings.Cut(got[i], "\t")
		if !ok {
			t.Errorf("line %d is not char TAB name: %q", i, got[i])
			continue
		}
		if char != string(r) {
			t.Errorf("line %d: alphabet says %q, Roles() says %q", i, char, string(r))
		}
		if name != RoleName(r) {
			t.Errorf("role %q: alphabet says %q, RoleName says %q", char, name, RoleName(r))
		}
	}
}

// The eleven, pinned. Adding a twelfth is a deliberate act and this is where
// it gets noticed -- a consumer that refuses a role outside the alphabet, as
// daffy does, starts refusing valid sprites the moment this grows.
func TestTheAlphabetIsEleven(t *testing.T) {
	if n := len(Roles()); n != 11 {
		t.Errorf("the alphabet is %d roles; every consumer's copy is of eleven", n)
	}
	seen := map[byte]bool{}
	for _, r := range Roles() {
		if seen[r] {
			t.Errorf("role %q appears twice", string(r))
		}
		seen[r] = true
	}
}

// The committed file is what the code produces.
//
// auklet/ALPHABET is the artifact another repository vendors, so it must not
// be allowed to go stale: a consumer testing against a file that no longer
// matches this package is worse off than one testing against nothing, because
// it will believe it has checked.
//
// Regenerate with: go run ./cmd/alphabet > auklet/ALPHABET
func TestAlphabetFileIsCurrent(t *testing.T) {
	body, err := os.ReadFile("ALPHABET")
	if err != nil {
		t.Fatalf("the published alphabet is missing: %v", err)
	}
	if string(body) != Alphabet() {
		t.Error("auklet/ALPHABET is stale; regenerate with: go run ./cmd/alphabet > auklet/ALPHABET")
	}
}
