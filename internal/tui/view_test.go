package tui

import (
	"strings"
	"testing"

	lipgloss "charm.land/lipgloss/v2"
)

func TestExecIndentFollowsParentHeadingLevel(t *testing.T) {
	doc := ParseMarkdown("# A\n## B\n### C\n```bash\necho hi\n```\n")
	m := NewModel(doc, "test.md")
	m.SetViewport(100, 20)

	// Outline is: h1, h2, h3, exec
	item := doc.Outline[3]
	indent := m.execIndent(item)
	if want := strings.Repeat("  ", 3); indent != want {
		t.Fatalf("exec indent = %q, want %q", indent, want)
	}
}

func TestExecIndentIsFlatInExecOnlyMode(t *testing.T) {
	doc := ParseMarkdown("# A\n## B\n### C\n```bash\necho hi\n```\n")
	m := NewModel(doc, "test.md")
	m.SetViewport(100, 20)
	m.execOnly = true

	item := doc.Outline[3]
	indent := m.execIndent(item)
	if indent != "" {
		t.Fatalf("exec indent in exec-only mode = %q, want empty", indent)
	}
}

func TestConstrainViewportClipsWidthAndHeight(t *testing.T) {
	in := strings.Join([]string{
		"0123456789abcdef",
		"line2",
		"line3",
		"line4",
	}, "\n")
	out := constrainViewport(in, 8, 3)
	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Fatalf("line count = %d, want 3", len(lines))
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w != 8 {
			t.Fatalf("line %d width = %d, want 8", i, w)
		}
	}
}
