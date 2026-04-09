package shared

import (
	"strings"
	"testing"
)

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

type testItem struct {
	id   string
	name string
}

func extract(t testItem) (string, string) { return t.id, t.name }

func TestMatchByName_ExactMatch(t *testing.T) {
	items := []testItem{{"id-1", "Alpha"}, {"id-2", "Beta"}}
	item, err := MatchByName("thing", "Alpha", items, extract)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.id != "id-1" {
		t.Errorf("expected id-1, got %s", item.id)
	}
}

func TestMatchByName_CaseInsensitive(t *testing.T) {
	items := []testItem{{"id-1", "Alpha"}}
	item, err := MatchByName("thing", "alpha", items, extract)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.id != "id-1" {
		t.Errorf("expected id-1, got %s", item.id)
	}
}

func TestMatchByName_SubstringMatch(t *testing.T) {
	items := []testItem{{"id-1", "Team Alpha"}}
	item, err := MatchByName("thing", "Alpha", items, extract)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.id != "id-1" {
		t.Errorf("expected id-1, got %s", item.id)
	}
}

func TestMatchByName_AmbiguousExact(t *testing.T) {
	items := []testItem{{"id-1", "Alpha"}, {"id-2", "Alpha"}}
	_, err := MatchByName("thing", "Alpha", items, extract)
	if err == nil {
		t.Fatal("expected ambiguous error")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("expected ambiguous error, got %q", err.Error())
	}
}

func TestMatchByName_AmbiguousSubstring(t *testing.T) {
	items := []testItem{{"id-1", "Team Alpha"}, {"id-2", "Team Beta"}}
	_, err := MatchByName("thing", "Team", items, extract)
	if err == nil {
		t.Fatal("expected ambiguous error")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("expected ambiguous error, got %q", err.Error())
	}
}

func TestMatchByName_NotFound(t *testing.T) {
	items := []testItem{{"id-1", "Alpha"}}
	_, err := MatchByName("thing", "Gamma", items, extract)
	if err == nil {
		t.Fatal("expected not-found error")
	}
	if !strings.Contains(err.Error(), "no thing found") {
		t.Errorf("expected not-found error, got %q", err.Error())
	}
}

func TestMatchByName_EmptyList(t *testing.T) {
	_, err := MatchByName("thing", "Alpha", []testItem{}, extract)
	if err == nil {
		t.Fatal("expected not-found error")
	}
}

func TestMatchByName_ExactTakesPrecedence(t *testing.T) {
	items := []testItem{{"id-sub", "Team Alpha"}, {"id-exact", "Alpha"}}
	item, err := MatchByName("thing", "Alpha", items, extract)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.id != "id-exact" {
		t.Errorf("expected exact match id-exact, got %s", item.id)
	}
}
