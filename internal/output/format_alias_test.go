package output

import "testing"

// TestParseFormatLenient pins the family's lenient parser behaviour now that
// ParseFormat is aliased to lib-agent-output: "ndjson"/"yml" aliases and
// case-insensitive input resolve, while genuinely unknown formats still error
// (fixable_by:agent). This is intentionally looser than the old local switch.
func TestParseFormatLenient(t *testing.T) {
	cases := []struct {
		in      string
		want    Format
		wantErr bool
	}{
		{"json", FormatJSON, false},
		{"JSON", FormatJSON, false},
		{"yaml", FormatYAML, false},
		{"yml", FormatYAML, false},
		{"YML", FormatYAML, false},
		{"jsonl", FormatNDJSON, false},
		{"ndjson", FormatNDJSON, false},
		{"NDJSON", FormatNDJSON, false},
		{"  json  ", FormatJSON, false},
		{"xml", "", true},
		{"", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseFormat(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tc.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("ParseFormat(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
