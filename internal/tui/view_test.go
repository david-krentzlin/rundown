package tui

import (
	"strings"
	"testing"
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
