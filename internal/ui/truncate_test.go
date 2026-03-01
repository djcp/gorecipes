package ui

import "testing"

func TestTruncate(t *testing.T) {
	tests := []struct {
		name string
		s    string
		max  int
		want string
	}{
		// Negative and zero max must never panic.
		{"negative max", "hello", -10, ""},
		{"zero max", "hello", 0, ""},

		// Strings that fit.
		{"exact fit", "hello", 5, "hello"},
		{"shorter than max", "hi", 10, "hi"},
		{"empty string", "", 5, ""},

		// Truncation with ellipsis.
		{"long string", "hello world", 8, "hello w…"},

		// Small max values (≤3 — no room for ellipsis).
		{"max 1", "hello", 1, "h"},
		{"max 2", "hello", 2, "he"},
		{"max 3", "hello", 3, "hel"},

		// Unicode.
		{"unicode exact", "héllo", 5, "héllo"},
		{"unicode truncate", "héllo world", 8, "héllo w…"},
		{"unicode negative", "héllo", -1, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := truncate(tc.s, tc.max)
			if got != tc.want {
				t.Errorf("truncate(%q, %d) = %q; want %q", tc.s, tc.max, got, tc.want)
			}
		})
	}
}
