// Package buildinfo answers "what is this binary and how old is it".
//
// Jane, on the tools in this estate: "seems like everyone needs to respond to
// --age in the enclave", and separately that bandersnatch and the auklet felt
// ABANDONED. The second is what this is for. Measured, none of these trees
// were stale -- every one had been committed to that day -- so the feeling was
// not about staleness, it was that there was no way to ASK. A binary that can
// say what commit it is and when it was put there turns "is this abandoned"
// from a feeling into a fact you get by typing four characters.
//
// Nothing here needs a build system. Go stamps the revision, the commit time
// and whether the tree was dirty into every binary for free, and the install
// time is the file's own modification time. A missing field is reported as
// missing rather than guessed at: a tool that invents a version is worse than
// one that admits it does not know.
package buildinfo

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

// Info is what a binary can say about itself without being told.
type Info struct {
	Revision  string // vcs.revision, empty when built outside a repository
	Committed time.Time
	Dirty     bool // vcs.modified: the tree had uncommitted changes
	Go        string
	Path      string
	Installed time.Time
}

// Read gathers it. Every field is best-effort and an absent one stays zero.
func Read() Info {
	var i Info
	if b, ok := debug.ReadBuildInfo(); ok {
		i.Go = b.GoVersion
		for _, s := range b.Settings {
			switch s.Key {
			case "vcs.revision":
				i.Revision = s.Value
			case "vcs.time":
				i.Committed, _ = time.Parse(time.RFC3339, s.Value)
			case "vcs.modified":
				i.Dirty = s.Value == "true"
			}
		}
	}
	if p, err := os.Executable(); err == nil {
		i.Path = p
		if st, err := os.Stat(p); err == nil {
			i.Installed = st.ModTime()
		}
	}
	return i
}

// Age is the report, in the shape puffin --age already uses so that reading
// two of these side by side needs no translation.
func (i Info) Age() string {
	var b strings.Builder
	rev := i.Revision
	if rev == "" {
		rev = "unstamped (built outside a repository)"
	} else if len(rev) > 7 {
		rev = rev[:7]
	}
	// the DIRTY flag is reported rather than swallowed. A binary built from a
	// modified tree is not the commit it names, and a tool that says a commit
	// it does not exactly match is the failure this whole idea exists to catch.
	if i.Dirty {
		rev += " (built from a modified tree)"
	}
	fmt.Fprintf(&b, "commit    %s", rev)
	if !i.Committed.IsZero() {
		fmt.Fprintf(&b, ", %s", since(i.Committed))
	}
	b.WriteString("\n")
	if !i.Installed.IsZero() {
		fmt.Fprintf(&b, "installed %s, %s\n", i.Installed.Local().Format("15:04:05"), since(i.Installed))
	}
	if i.Path != "" {
		fmt.Fprintf(&b, "path      %s\n", i.Path)
	}
	if i.Go != "" {
		fmt.Fprintf(&b, "go        %s\n", i.Go)
	}
	return b.String()
}

// since renders an age the way a person says it: the largest unit and no more.
func since(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours())/24)
}
