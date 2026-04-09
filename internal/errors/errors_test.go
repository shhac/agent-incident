package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("something broke", FixableByHuman)
	if err.Message != "something broke" {
		t.Fatalf("expected message %q, got %q", "something broke", err.Message)
	}
	if err.FixableBy != FixableByHuman {
		t.Fatalf("expected fixable_by %q, got %q", FixableByHuman, err.FixableBy)
	}
	if err.Hint != "" {
		t.Fatalf("expected empty hint, got %q", err.Hint)
	}
	if err.Cause != nil {
		t.Fatal("expected nil cause")
	}
}

func TestNewf(t *testing.T) {
	err := Newf(FixableByAgent, "bad input %q for field %s", "abc", "name")
	want := `bad input "abc" for field name`
	if err.Message != want {
		t.Fatalf("expected message %q, got %q", want, err.Message)
	}
	if err.FixableBy != FixableByAgent {
		t.Fatalf("expected fixable_by %q, got %q", FixableByAgent, err.FixableBy)
	}
}

func TestWrap(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		if got := Wrap(nil, FixableByRetry); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})

	t.Run("wraps error", func(t *testing.T) {
		orig := fmt.Errorf("disk full")
		wrapped := Wrap(orig, FixableByRetry)
		if wrapped.Message != "disk full" {
			t.Fatalf("expected message %q, got %q", "disk full", wrapped.Message)
		}
		if wrapped.FixableBy != FixableByRetry {
			t.Fatalf("expected fixable_by %q, got %q", FixableByRetry, wrapped.FixableBy)
		}
		if wrapped.Cause != orig {
			t.Fatal("expected cause to be the original error")
		}
	})
}

func TestWithHint(t *testing.T) {
	err := New("auth failed", FixableByHuman).WithHint("check your API key")
	if err.Hint != "check your API key" {
		t.Fatalf("expected hint %q, got %q", "check your API key", err.Hint)
	}
}

func TestWithCause(t *testing.T) {
	cause := fmt.Errorf("underlying")
	err := New("wrapper", FixableByAgent).WithCause(cause)
	if err.Cause != cause {
		t.Fatal("expected cause to match")
	}
}

func TestError(t *testing.T) {
	err := New("test message", FixableByAgent)
	if err.Error() != "test message" {
		t.Fatalf("expected %q, got %q", "test message", err.Error())
	}
}

func TestUnwrap(t *testing.T) {
	t.Run("with cause", func(t *testing.T) {
		cause := fmt.Errorf("root")
		err := New("outer", FixableByAgent).WithCause(cause)
		if err.Unwrap() != cause {
			t.Fatal("Unwrap should return cause")
		}
	})

	t.Run("without cause", func(t *testing.T) {
		err := New("outer", FixableByAgent)
		if err.Unwrap() != nil {
			t.Fatal("Unwrap should return nil when no cause")
		}
	})
}

func TestAs(t *testing.T) {
	cause := New("inner", FixableByHuman).WithHint("try again")
	wrapped := fmt.Errorf("outer: %w", cause)

	var target *APIError
	if !As(wrapped, &target) {
		t.Fatal("As should find APIError in chain")
	}
	if target.Hint != "try again" {
		t.Fatalf("expected hint %q, got %q", "try again", target.Hint)
	}
}

func TestAsNotFound(t *testing.T) {
	plain := errors.New("plain error")
	var target *APIError
	if As(plain, &target) {
		t.Fatal("As should return false for non-APIError")
	}
}
