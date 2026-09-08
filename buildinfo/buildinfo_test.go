package buildinfo

import (
	"strings"
	"testing"
	"time"
)

// An absent field is reported as absent, never guessed at.
//
// The whole point is answering "what is this and how old is it" truthfully, so
// a binary built outside a repository has to say so rather than print a
// plausible blank. A tool that invents a version is worse than one that admits
// it does not know, because the invented one gets believed.
func TestUnstampedSaysSo(t *testing.T) {
	got := Info{Go: "go1.26.4", Path: "/tmp/x"}.Age()
	if !strings.Contains(got, "unstamped") {
		t.Errorf("a build with no revision does not say so:\n%s", got)
	}
	if strings.Contains(got, "installed") {
		t.Errorf("a zero install time was reported anyway:\n%s", got)
	}
}

// A binary built from a modified tree is not the commit it names, and saying
// the commit without the caveat is the failure this package exists to catch.
func TestDirtyIsNotSwallowed(t *testing.T) {
	clean := Info{Revision: "1f55b35abcdef"}.Age()
	dirty := Info{Revision: "1f55b35abcdef", Dirty: true}.Age()
	if strings.Contains(clean, "modified") {
		t.Errorf("a clean build claims it was modified:\n%s", clean)
	}
	if !strings.Contains(dirty, "modified") {
		t.Errorf("a dirty build does not say so:\n%s", dirty)
	}
	if !strings.Contains(clean, "1f55b35") || strings.Contains(clean, "abcdef") {
		t.Errorf("the revision is not shortened to seven:\n%s", clean)
	}
}

// The age reads in the largest unit that fits, because "5400s ago" is a number
// and "1h ago" is an answer.
func TestSinceReadsLikeAPerson(t *testing.T) {
	for _, c := range []struct {
		ago  time.Duration
		want string
	}{
		{30 * time.Second, "30s ago"},
		{90 * time.Minute, "1h ago"},
		{50 * time.Hour, "2d ago"},
	} {
		if got := since(time.Now().Add(-c.ago)); got != c.want {
			t.Errorf("%v ago reads %q, want %q", c.ago, got, c.want)
		}
	}
}

// Read is best-effort and must not panic on a machine that tells it nothing.
func TestReadIsSafe(t *testing.T) {
	if got := Read().Age(); got == "" {
		t.Error("Read().Age() said nothing at all")
	}
}
