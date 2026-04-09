package shared

import "testing"

func TestLooksLikeID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"01HRDKWWNGX330JQ4J1PERJP8Y", true},
		{"01GF688HF8NVA17SNPGPYV5XNZ", true},
		{"Engineering", false},
		{"alice@example.com", false},
		{"Primary On-Call", false},
		{"", false},
		{"short", false},
		// I, L, O, U are not in the ULID charset
		{"01HRDKWWNGX330JQ4J1IIIIII", false},
		{"01HRDKWWNGX330JQ4J1LLLLLL", false},
		{"01HRDKWWNGX330JQ4J1OOOOOO", false},
		{"01HRDKWWNGX330JQ4J1UUUUUU", false},
	}

	for _, tt := range tests {
		if got := looksLikeID(tt.input); got != tt.want {
			t.Errorf("looksLikeID(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
