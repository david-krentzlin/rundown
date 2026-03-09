package tui

import (
	"fmt"
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

func TestClipLineForcesSingleLine(t *testing.T) {
	out := clipLine("aaaaaaaaaaaaaaaa\nafter", 8)
	if strings.Contains(out, "\n") {
		t.Fatalf("clipLine produced multiline output: %q", out)
	}
	if w := lipgloss.Width(out); w != 8 {
		t.Fatalf("clipLine width = %d, want 8", w)
	}
}

func TestRenderLogPanelKeepsFixedHeightWhenScrollable(t *testing.T) {
	doc := ParseMarkdown("# A\n```bash\necho hi\n```\n")
	m := NewModel(doc, "test.md")
	m.execPanelVisible = true
	m.SetViewport(50, 20)

	execIdx := -1
	for i, item := range m.doc.Outline {
		if item.Kind == NodeExec {
			execIdx = i
			break
		}
	}
	if execIdx < 0 {
		t.Fatal("expected executable item")
	}
	blockID := m.doc.Outline[execIdx].ID

	logs := make([]ExecLogLine, 0, 40)
	for i := 0; i < 40; i++ {
		logs = append(logs, ExecLogLine{Text: fmt.Sprintf("log-%02d %s", i, strings.Repeat("x", 120)), Kind: "output"})
	}
	logs[5] = ExecLogLine{Text: "first line\nsecond line that must not expand panel rows", Kind: "output"}
	m.execHistory[blockID] = []ExecRecord{{
		Title:  "bash block",
		Lang:   "bash",
		Status: "completed (0)",
		Logs:   logs,
	}}
	m.execViewIndex[blockID] = 0
	m.outlineIdx = execIdx
	m.execLogScroll = 17

	panelH := 10
	panel := m.renderLogPanel(panelH)
	lines := strings.Split(panel, "\n")
	if len(lines) != panelH {
		t.Fatalf("panel lines = %d, want %d", len(lines), panelH)
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w != m.width {
			t.Fatalf("panel line %d width = %d, want %d", i, w, m.width)
		}
	}
}

func TestRenderExecLogLineStylesStderr(t *testing.T) {
	m := NewModel(ParseMarkdown("# A\n"), "test.md")
	line := m.renderExecLogLine(ExecLogLine{Text: "boom", Stream: "stderr", Kind: "output"}, 20)
	if !strings.Contains(line, "\x1b[") {
		t.Fatalf("expected styled stderr line, got %q", line)
	}
}
