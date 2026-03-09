package app

import "testing"

func TestDisplayName(t *testing.T) {
	if got := DisplayName(); got != "rundown" {
		t.Fatalf("DisplayName() = %q, want %q", got, "rundown")
	}
}
