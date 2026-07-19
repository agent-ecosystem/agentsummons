package agentsummons

import "testing"

func TestVersionNewer(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"2.1.205", "2.1.204", true},
		{"2.1.204", "2.1.204", false},
		{"2.1.203", "2.1.204", false},
		{"2.2", "2.1.9", true},
		{"2.1", "2.1.0", false},
		{"2.1.0", "2.1", false},
		{"10.0.0", "9.9.9", true},
		{"1.0.0-beta", "1.0.0", false},
		// Unparseable segments end the comparison as equal: never a false
		// drift signal from a malformed version.
		{"abc", "1.0", false},
		{"1.x", "1.0", false},
		// Segments too large for int also end the comparison as equal.
		{"99999999999999999999", "1.0", false},
		{"1.0", "99999999999999999999", false},
	}
	for _, tc := range cases {
		if got := VersionNewer(tc.a, tc.b); got != tc.want {
			t.Errorf("VersionNewer(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}
