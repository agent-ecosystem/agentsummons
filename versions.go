package agentsummons

import (
	"regexp"
	"strconv"
	"strings"
)

// LastValidated records the newest release of each harness whose flag
// surface the spec table in harnesses.go was validated against. It is the
// invocation-side counterpart of the agentminutes table, which covers the
// transcript format surface — the two drift independently, so the tables
// are deliberately separate. Like its counterpart it is coverage
// documentation, not a compatibility bound: flags usually survive newer
// releases, and a removed flag fails loudly at run time anyway.
// Alphabetical; update an entry when the spec table is re-checked against
// a newer release (see DEVELOPMENT.md). Callers must treat the map as
// read-only.
var LastValidated = map[ID]string{
	Antigravity: "1.1.8",
	ClaudeCode:  "2.1.212",
	Codex:       "0.146.0",
}

// versionRe extracts the first dotted version from version-command output.
var versionRe = regexp.MustCompile(`\d+(?:\.\d+)+`)

// VersionNewer reports whether version a is strictly newer than b,
// comparing dot-separated segments numerically by their leading digits
// (missing segments count as zero). It is deliberately lenient: harness
// versions are not guaranteed semver, and an unparseable segment ends the
// comparison as equal, so a malformed version never triggers a false
// drift hint. Copied from agentminutes harness/versions.go — importing it
// would invert the dependency direction between the two libraries.
func VersionNewer(a, b string) bool {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		av, aok := segmentNumber(as, i)
		bv, bok := segmentNumber(bs, i)
		if !aok || !bok {
			return false
		}
		if av != bv {
			return av > bv
		}
	}
	return false
}

// segmentNumber returns the numeric value of the i-th version segment: the
// segment's leading digits (so "0-alpha" reads as 0), or zero when past the
// end. ok is false when the segment exists but starts with no digit.
func segmentNumber(segs []string, i int) (n int, ok bool) {
	if i >= len(segs) {
		return 0, true
	}
	s := segs[i]
	j := 0
	for j < len(s) && s[j] >= '0' && s[j] <= '9' {
		j++
	}
	if j == 0 {
		return 0, false
	}
	v, err := strconv.Atoi(s[:j])
	if err != nil {
		return 0, false
	}
	return v, true
}
