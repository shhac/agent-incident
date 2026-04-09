package shared

import (
	"testing"
	"time"

	agenterrors "github.com/shhac/agent-incident/internal/errors"
)

func TestParseTimeEmpty(t *testing.T) {
	got, err := ParseTime("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.IsZero() {
		t.Fatalf("expected zero time, got %v", got)
	}
}

func TestParseTimeNow(t *testing.T) {
	before := time.Now()
	got, err := ParseTime("now")
	after := time.Now()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Before(before) || got.After(after) {
		t.Fatalf("expected time between %v and %v, got %v", before, after, got)
	}
}

func TestParseTimeRelativeBackward(t *testing.T) {
	tests := []struct {
		input    string
		minDelta time.Duration
		maxDelta time.Duration
	}{
		{"now-15m", 14*time.Minute + 50*time.Second, 15*time.Minute + 10*time.Second},
		{"now-1h", 59*time.Minute + 50*time.Second, 60*time.Minute + 10*time.Second},
		{"now-1d", 23*time.Hour + 59*time.Minute + 50*time.Second, 24*time.Hour + 10*time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			now := time.Now()
			got, err := ParseTime(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			delta := now.Sub(got)
			if delta < tt.minDelta || delta > tt.maxDelta {
				t.Fatalf("expected delta between %v and %v, got %v", tt.minDelta, tt.maxDelta, delta)
			}
		})
	}
}

func TestParseTimeRelativeForward(t *testing.T) {
	tests := []struct {
		input    string
		minDelta time.Duration
		maxDelta time.Duration
	}{
		{"now+1h", 59*time.Minute + 50*time.Second, 60*time.Minute + 10*time.Second},
		{"now+30m", 29*time.Minute + 50*time.Second, 30*time.Minute + 10*time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			now := time.Now()
			got, err := ParseTime(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			delta := got.Sub(now)
			if delta < tt.minDelta || delta > tt.maxDelta {
				t.Fatalf("expected delta between %v and %v, got %v", tt.minDelta, tt.maxDelta, delta)
			}
		})
	}
}

func TestParseTimeRFC3339(t *testing.T) {
	input := "2024-06-15T14:30:00Z"
	got, err := ParseTime(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
	if !got.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestParseTimeUnixEpoch(t *testing.T) {
	got, err := ParseTime("1718457000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Unix(1718457000, 0)
	if !got.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestParseTimeInvalid(t *testing.T) {
	inputs := []string{
		"yesterday",
		"not-a-time",
		"2024/06/15",
		"now?5m",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			_, err := ParseTime(input)
			if err == nil {
				t.Fatalf("expected error for %q", input)
			}
			var apiErr *agenterrors.APIError
			if !agenterrors.As(err, &apiErr) {
				t.Fatalf("expected APIError, got %T", err)
			}
		})
	}
}

func TestParseTimeRangeDefaults(t *testing.T) {
	before := time.Now()
	from, to, err := ParseTimeRange("", "")
	after := time.Now()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// "from" should default to ~1 hour ago
	expectedFrom := before.Add(-1 * time.Hour)
	if from.Before(expectedFrom.Add(-5*time.Second)) || from.After(after.Add(-1*time.Hour).Add(5*time.Second)) {
		t.Fatalf("from should be ~1h ago, got %v", from)
	}

	// "to" should default to ~now
	if to.Before(before.Add(-5*time.Second)) || to.After(after.Add(5*time.Second)) {
		t.Fatalf("to should be ~now, got %v", to)
	}
}

func TestParseTimeRangeWithValues(t *testing.T) {
	from, to, err := ParseTimeRange("2024-01-01T00:00:00Z", "2024-01-02T00:00:00Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedTo := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	if !from.Equal(expectedFrom) {
		t.Fatalf("expected from %v, got %v", expectedFrom, from)
	}
	if !to.Equal(expectedTo) {
		t.Fatalf("expected to %v, got %v", expectedTo, to)
	}
}

func TestParseTimeRangeInvalidFrom(t *testing.T) {
	_, _, err := ParseTimeRange("garbage", "")
	if err == nil {
		t.Fatal("expected error for invalid from")
	}
}

func TestParseTimeRangeInvalidTo(t *testing.T) {
	_, _, err := ParseTimeRange("", "garbage")
	if err == nil {
		t.Fatal("expected error for invalid to")
	}
}
