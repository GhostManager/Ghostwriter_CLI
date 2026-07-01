package internal

import "testing"

func TestCommandOutputTailKeepsBoundedTail(t *testing.T) {
	tail := newCommandOutputTail(10)

	if _, err := tail.Write([]byte("first-")); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	if _, err := tail.Write([]byte("second")); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	if got := tail.String(); got != "rst-second" {
		t.Fatalf("expected bounded tail %q, got %q", "rst-second", got)
	}
}
