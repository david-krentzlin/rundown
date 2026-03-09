package tui

import "testing"

func TestConcealStateAtInsideFence(t *testing.T) {
	lines := []string{
		"# Title",
		"```bash",
		"echo one",
		"echo two",
		"```",
		"after",
	}

	st := concealStateAt(lines, 3) // viewport starts on "echo two"
	if !st.inFence {
		t.Fatal("expected inFence=true when viewport starts inside fenced block")
	}

	st = concealStateAt(lines, 5) // viewport starts after closing fence
	if st.inFence {
		t.Fatal("expected inFence=false after fenced block is closed")
	}
}
