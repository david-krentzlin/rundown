package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestModelSyncMarkdownToOutline(t *testing.T) {
	doc := ParseMarkdown("# A\ntext\n## B\n```go\nfmt.Println(1)\n```\n")
	m := NewModel(doc, "test.md")
	m.SetViewport(100, 20)

	m.cursorLine = doc.Outline[2].Line
	m.syncOutlineFromMarkdown()

	if got := m.outlineIdx; got != 2 {
		t.Fatalf("outlineIdx = %d, want 2", got)
	}

	m.cursorLine = 6 // below the code block
	m.syncOutlineFromMarkdown()
	if got := m.outlineIdx; got != 1 {
		t.Fatalf("outlineIdx after leaving exec block = %d, want 1", got)
	}
}

func TestModelSyncOutlineToMarkdown(t *testing.T) {
	doc := ParseMarkdown("# A\ntext\n## B\nmore\n")
	m := NewModel(doc, "test.md")
	m.SetViewport(80, 10)

	m.outlineIdx = 1
	m.syncMarkdownFromOutline()

	if got, want := m.cursorLine, doc.Outline[1].Line; got != want {
		t.Fatalf("cursorLine = %d, want %d", got, want)
	}
	if got, want := m.mdTop, 0; got != want {
		t.Fatalf("mdTop = %d, want %d", got, want)
	}
}

func TestToggleExecutableOnly(t *testing.T) {
	doc := ParseMarkdown("# A\n## B\n```bash\necho ok\n```\n")
	m := NewModel(doc, "test.md")

	m.execOnly = true
	items := m.visibleOutlineIndexes()
	if len(items) != 1 {
		t.Fatalf("visible outline items = %d, want 1", len(items))
	}
	if doc.Outline[items[0]].Kind != NodeExec {
		t.Fatalf("visible item kind = %v, want NodeExec", doc.Outline[items[0]].Kind)
	}
}

func TestMarkdownHJKLHeadingNavigation(t *testing.T) {
	doc := ParseMarkdown("# Root\ntext\n## Child\n### Leaf\n## Child Two\n")
	m := NewModel(doc, "test.md")
	m.SetViewport(100, 20)

	// J: next heading from current cursor
	m.cursorLine = 0
	m.handleMarkdownKey("J")
	if got, want := m.cursorLine, doc.Outline[1].Line; got != want {
		t.Fatalf("J cursorLine = %d, want %d", got, want)
	}

	// L: first child heading
	m.handleMarkdownKey("L")
	if got, want := m.cursorLine, doc.Outline[2].Line; got != want {
		t.Fatalf("L cursorLine = %d, want %d", got, want)
	}

	// H: parent heading
	m.handleMarkdownKey("H")
	if got, want := m.cursorLine, doc.Outline[1].Line; got != want {
		t.Fatalf("H cursorLine = %d, want %d", got, want)
	}

	// K: previous heading in document order
	m.handleMarkdownKey("K")
	if got, want := m.cursorLine, doc.Outline[0].Line; got != want {
		t.Fatalf("K cursorLine = %d, want %d", got, want)
	}

	// H: parent heading, not just previous heading in order.
	m.cursorLine = doc.Outline[3].Line // Child Two
	m.syncOutlineFromMarkdown()
	m.handleMarkdownKey("H")
	if got, want := m.cursorLine, doc.Outline[0].Line; got != want {
		t.Fatalf("H (parent) cursorLine = %d, want %d", got, want)
	}
}

func TestCollapseCurrentHidesDescendantsButKeepsSiblingVisible(t *testing.T) {
	doc := ParseMarkdown("# A\n## B\n### C\n## D\n")
	m := NewModel(doc, "test.md")
	m.SetViewport(80, 12)

	// focus heading B and collapse it
	m.outlineIdx = 1
	m.collapseCurrent()

	visible := m.visibleOutlineIndexes()
	want := []int{0, 1, 3}
	if len(visible) != len(want) {
		t.Fatalf("visible len = %d, want %d (%v)", len(visible), len(want), visible)
	}
	for i := range want {
		if visible[i] != want[i] {
			t.Fatalf("visible[%d] = %d, want %d (all=%v)", i, visible[i], want[i], visible)
		}
	}
}

func TestOutlineUsesJKForUpDown(t *testing.T) {
	doc := ParseMarkdown("# A\n## B\n## C\n")
	m := NewModel(doc, "test.md")
	m.focus = PaneOutline
	m.SetViewport(80, 12)

	m.outlineIdx = 0
	m.handleOutlineKey("j")
	if got := m.outlineIdx; got != 1 {
		t.Fatalf("j outlineIdx = %d, want 1", got)
	}

	m.handleOutlineKey("k")
	if got := m.outlineIdx; got != 0 {
		t.Fatalf("k outlineIdx = %d, want 0", got)
	}
}

func TestMouseWheelScrollsMarkdownPane(t *testing.T) {
	lines := make([]string, 0, 120)
	for i := 0; i < 120; i++ {
		lines = append(lines, fmt.Sprintf("line %d", i))
	}
	doc := ParseMarkdown(strings.Join(lines, "\n"))
	m := NewModel(doc, "test.md")
	m.SetViewport(100, 20)

	m.handleMouseWheel(tea.MouseWheelMsg(tea.Mouse{X: 2, Y: 2, Button: tea.MouseWheelDown}))
	if m.mdTop == 0 {
		t.Fatal("expected markdown top to move on wheel down")
	}

	before := m.mdTop
	// Right pane x-position should not scroll markdown.
	m.handleMouseWheel(tea.MouseWheelMsg(tea.Mouse{X: 90, Y: 2, Button: tea.MouseWheelDown}))
	if m.mdTop != before {
		t.Fatalf("mdTop changed from %d to %d when wheel event was outside markdown pane", before, m.mdTop)
	}
}

func TestSyncOutlineToMarkdownSetsTopWithContext(t *testing.T) {
	doc := ParseMarkdown("# A\n1\n2\n3\n4\n5\n## B\nline\n")
	m := NewModel(doc, "test.md")
	m.SetViewport(80, 8)
	m.outlineIdx = 1 // heading B at line 6

	m.syncMarkdownFromOutline()

	if got, want := m.cursorLine, 6; got != want {
		t.Fatalf("cursorLine = %d, want %d", got, want)
	}
	if got, want := m.mdTop, 1; got != want {
		t.Fatalf("mdTop = %d, want %d", got, want)
	}
}

func TestSelectedLineRangeForHeadingSection(t *testing.T) {
	doc := ParseMarkdown("# A\nintro\n## B\nb text\n## C\nc text\n")
	m := NewModel(doc, "test.md")
	m.SetViewport(80, 12)
	m.outlineIdx = 0 // heading A

	start, end, ok := m.selectedLineRange()
	if !ok {
		t.Fatal("expected selected range")
	}
	if start != 0 || end != len(doc.Lines)-1 {
		t.Fatalf("heading range = [%d,%d], want [0,%d]", start, end, len(doc.Lines)-1)
	}
}

func TestSelectedLineRangeForExecBlock(t *testing.T) {
	doc := ParseMarkdown("# A\n```bash\necho hi\n```\ntext\n")
	m := NewModel(doc, "test.md")
	m.SetViewport(80, 12)
	m.outlineIdx = 1 // bash block

	start, end, ok := m.selectedLineRange()
	if !ok {
		t.Fatal("expected selected range")
	}
	if start != 1 || end != 3 {
		t.Fatalf("exec range = [%d,%d], want [1,3]", start, end)
	}
}

func TestTabSwitchesPaneOnKeyPress(t *testing.T) {
	doc := ParseMarkdown("# A\n")
	m := NewModel(doc, "test.md")

	if m.focus != PaneMarkdown {
		t.Fatalf("initial focus = %v, want markdown", m.focus)
	}

	m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
	if m.focus != PaneOutline {
		t.Fatalf("focus after tab key press = %v, want outline", m.focus)
	}

	m.handleKey("ctrl+i")
	if m.focus != PaneMarkdown {
		t.Fatalf("focus after ctrl+i key press = %v, want markdown", m.focus)
	}
}

func TestEnableGlamourMessageTurnsOnMarkdownRenderer(t *testing.T) {
	doc := ParseMarkdown("# A\ntext\n")
	m := NewModel(doc, "test.md")
	if m.useGlamour {
		t.Fatal("expected glamour to be disabled initially")
	}

	m.Update(enableGlamourMsg{})
	if !m.useGlamour {
		t.Fatal("expected glamour to be enabled after async message")
	}
}
