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

	if m.focus != PaneOutline {
		t.Fatalf("initial focus = %v, want outline", m.focus)
	}

	m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
	if m.focus != PaneMarkdown {
		t.Fatalf("focus after tab key press = %v, want markdown", m.focus)
	}

	m.handleKey("ctrl+i")
	if m.focus != PaneOutline {
		t.Fatalf("focus after ctrl+i key press = %v, want outline", m.focus)
	}
}

func TestModelInitDoesNotAutoExecute(t *testing.T) {
	doc := ParseMarkdown("# A\ntext\n")
	m := NewModel(doc, "test.md")
	_, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	_ = cmd
	if m.execRunning {
		t.Fatal("execution should not start on init")
	}
	if len(m.execHistory) != 0 {
		t.Fatal("execution history should be empty on load")
	}
}

func TestLogPanelHeightFitsViewport(t *testing.T) {
	doc := ParseMarkdown("# A\n")
	m := NewModel(doc, "test.md")
	m.execPanelVisible = true
	m.SetViewport(80, 14)

	main := m.mainHeight()
	panel := m.logPanelHeight()
	if main < 3 {
		t.Fatalf("main height = %d, want >= 3", main)
	}
	if panel < 3 {
		t.Fatalf("panel height = %d, want >= 3", panel)
	}
	if main+panel > m.height-2 {
		t.Fatalf("layout overflow: main(%d)+panel(%d) > available(%d)", main, panel, m.height-2)
	}
}

func TestRenderedViewFitsViewportHeightWithPanel(t *testing.T) {
	doc := ParseMarkdown("# A\n```bash\necho hi\n```\n")
	m := NewModel(doc, "test.md")
	m.execPanelVisible = true
	m.SetViewport(100, 24)

	v := m.View()
	lines := countLines(v.Content)
	if lines > m.height {
		t.Fatalf("rendered lines = %d, viewport height = %d", lines, m.height)
	}
}

func TestNewModelSelectsFirstHeadingNotExec(t *testing.T) {
	doc := ParseMarkdown("```bash\necho hi\n```\n# H1\n")
	m := NewModel(doc, "test.md")
	if m.doc.Outline[m.outlineIdx].Kind != NodeHeading {
		t.Fatalf("initial outline kind = %v, want heading", m.doc.Outline[m.outlineIdx].Kind)
	}
}

func TestCtrlAAndCtrlENavigateDocument(t *testing.T) {
	doc := ParseMarkdown("# A\nline1\nline2\n## B\nline3\n")
	m := NewModel(doc, "test.md")
	m.SetViewport(80, 8)

	m.cursorLine = 4
	m.handleKey("ctrl+a")
	if m.cursorLine != 0 {
		t.Fatalf("ctrl+a cursorLine = %d, want 0", m.cursorLine)
	}

	m.handleKey("ctrl+e")
	if want := len(doc.Lines) - 1; m.cursorLine != want {
		t.Fatalf("ctrl+e cursorLine = %d, want %d", m.cursorLine, want)
	}
}

func TestExecHistoryAssociatesByBlockID(t *testing.T) {
	doc := ParseMarkdown("# A\n```bash\necho a\n```\n# B\n```bash\necho b\n```\n")
	m := NewModel(doc, "test.md")
	m.SetViewport(80, 20)

	var firstExec, secondExec int
	firstExec, secondExec = -1, -1
	for i, item := range m.doc.Outline {
		if item.Kind != NodeExec {
			continue
		}
		if firstExec < 0 {
			firstExec = i
		} else {
			secondExec = i
			break
		}
	}
	if firstExec < 0 || secondExec < 0 {
		t.Fatalf("expected two executable blocks, got first=%d second=%d", firstExec, secondExec)
	}

	firstID := m.doc.Outline[firstExec].ID
	secondID := m.doc.Outline[secondExec].ID
	m.execHistory[firstID] = []ExecRecord{{Title: "first", Lang: "bash", Status: "completed (0)", Logs: []ExecLogLine{{Text: "first log", Kind: "output"}}}}
	m.execHistory[secondID] = []ExecRecord{{Title: "second", Lang: "bash", Status: "failed (1)", Logs: []ExecLogLine{{Text: "second log", Kind: "output"}}}}
	m.execViewIndex[firstID] = 0
	m.execViewIndex[secondID] = 0

	// Reorder outline items to emulate index instability.
	m.doc.Outline[firstExec], m.doc.Outline[secondExec] = m.doc.Outline[secondExec], m.doc.Outline[firstExec]

	// Select the block that originally had secondID and ensure we get second log.
	selectedIdx := m.outlineIndexByID(secondID)
	if selectedIdx < 0 {
		t.Fatalf("expected to find outline item by id %q after reorder", secondID)
	}
	m.outlineIdx = selectedIdx
	rec, current, total, ok := m.execPanelData()
	if !ok {
		t.Fatal("expected exec panel data")
	}
	if rec.Title != "second" {
		t.Fatalf("title = %q, want second", rec.Title)
	}
	if len(rec.Logs) != 1 || rec.Logs[0].Text != "second log" {
		t.Fatalf("logs = %v, want [second log]", rec.Logs)
	}
	if rec.Status != "failed (1)" || current != 1 || total != 1 {
		t.Fatalf("status/current/total = %q/%d/%d, want failed (1)/1/1", rec.Status, current, total)
	}
}

func TestExecScrollClampsUsingFullVisibleLogArea(t *testing.T) {
	doc := ParseMarkdown("# A\n```bash\necho hi\n```\n")
	m := NewModel(doc, "test.md")
	m.execPanelVisible = true
	m.SetViewport(80, 24)

	execIdx := -1
	for i, item := range m.doc.Outline {
		if item.Kind == NodeExec {
			execIdx = i
			break
		}
	}
	if execIdx < 0 {
		t.Fatal("expected executable block in outline")
	}

	blockID := m.doc.Outline[execIdx].ID
	visible := max(1, m.logPanelHeight()-4)
	logs := make([]ExecLogLine, visible)
	for i := 0; i < visible; i++ {
		logs[i] = ExecLogLine{Text: fmt.Sprintf("line-%d", i), Kind: "output"}
	}
	m.execHistory[blockID] = []ExecRecord{{
		Title:  "bash block",
		Lang:   "bash",
		Status: "completed (0)",
		Logs:   logs,
	}}
	m.execViewIndex[blockID] = 0
	m.outlineIdx = execIdx

	m.execScroll(999)
	if m.execLogScroll != 0 {
		t.Fatalf("execLogScroll = %d, want 0 when logs exactly fit viewport", m.execLogScroll)
	}
}

func TestExecLineAutoScrollsWhenFollowingTail(t *testing.T) {
	doc := ParseMarkdown("# A\n```bash\necho hi\n```\n")
	m := NewModel(doc, "test.md")
	m.execPanelVisible = true
	m.SetViewport(80, 24)

	execIdx := -1
	for i, item := range m.doc.Outline {
		if item.Kind == NodeExec {
			execIdx = i
			break
		}
	}
	if execIdx < 0 {
		t.Fatal("expected executable block")
	}
	blockID := m.doc.Outline[execIdx].ID
	m.outlineIdx = execIdx
	m.execRunBlockID = blockID
	m.execViewBlockID = blockID
	m.execRunning = true
	m.execFollowTail = true
	visible := max(1, m.logPanelHeight()-4)
	logs := make([]ExecLogLine, visible+3)
	for i := range logs {
		logs[i] = ExecLogLine{Text: fmt.Sprintf("line-%d", i), Kind: "output"}
	}
	m.execHistory[blockID] = []ExecRecord{{
		Title:  "bash block",
		Lang:   "bash",
		Status: "running",
		Logs:   logs,
	}}
	m.execViewIndex[blockID] = 0
	m.execLogScroll = 0

	m.handleInternalMsg(execLineMsg{line: "next-line", stream: "stdout"})
	if m.execLogScroll == 0 {
		t.Fatalf("execLogScroll = %d, want >0 after auto-follow append", m.execLogScroll)
	}
}

func TestManualScrollUpDisablesTailFollow(t *testing.T) {
	doc := ParseMarkdown("# A\n```bash\necho hi\n```\n")
	m := NewModel(doc, "test.md")
	m.execPanelVisible = true
	m.SetViewport(80, 24)
	m.execFollowTail = true

	m.handleKey("pgup")
	if m.execFollowTail {
		t.Fatal("expected pgup to disable log tail follow mode")
	}
	m.handleKey("end")
	if !m.execFollowTail {
		t.Fatal("expected end to re-enable log tail follow mode")
	}
}

func TestQuitStopsRunningExecution(t *testing.T) {
	doc := ParseMarkdown("# A\n")
	m := NewModel(doc, "test.md")

	canceled := false
	m.execRunning = true
	m.execCancel = func() { canceled = true }

	quit, _, handled := m.handleGlobalKey("ctrl+q")
	if !handled {
		t.Fatal("expected ctrl+q to be handled")
	}
	if !quit {
		t.Fatal("expected ctrl+q to request quit")
	}
	if !canceled {
		t.Fatal("expected running execution to be canceled before quit")
	}
}

func TestCannotStartExecutionWhileAnotherRunIsActive(t *testing.T) {
	doc := ParseMarkdown("# A\n```bash\necho hi\n```\n")
	m := NewModel(doc, "test.md")
	m.SetViewport(80, 20)

	execIdx := -1
	for i, item := range m.doc.Outline {
		if item.Kind == NodeExec {
			execIdx = i
			break
		}
	}
	if execIdx < 0 {
		t.Fatal("expected executable block in outline")
	}

	m.outlineIdx = execIdx
	m.execRunning = true
	before := len(m.execHistory[m.doc.Outline[execIdx].ID])

	cmd := m.handleOutlineKey("r")
	if cmd != nil {
		t.Fatal("expected nil cmd when execution is already running")
	}
	after := len(m.execHistory[m.doc.Outline[execIdx].ID])
	if after != before {
		t.Fatalf("execution history changed while already running: before=%d after=%d", before, after)
	}
}

func TestHelpOverlayToggleAndKeyCapture(t *testing.T) {
	doc := ParseMarkdown("# A\nline\n")
	m := NewModel(doc, "test.md")
	m.SetViewport(80, 20)

	m.handleKey("?")
	if !m.helpVisible {
		t.Fatal("expected help to become visible after '?'")
	}

	before := m.cursorLine
	m.handleKey("j")
	if m.cursorLine != before {
		t.Fatal("expected non-help keys to be captured while help is visible")
	}

	m.handleKey("esc")
	if m.helpVisible {
		t.Fatal("expected help to close on esc")
	}
}

func countLines(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}
