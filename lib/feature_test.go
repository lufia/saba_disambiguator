package sabadisambiguator

import (
	"testing"
)

func TestRemoveScreenNames(t *testing.T) {
	tests := map[string]string{
		"@screen text":        "text",
		"@screen aaa@screen2": "aaa",
		".@screen":            ".",
	}
	for in, want := range tests {
		s := removeScreenNames(in)
		if s != want {
			t.Errorf("removeScreenNames(%q) = %q; want %q", in, s, want)
		}
	}
}
