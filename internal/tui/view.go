package tui

import (
	"fmt"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
)

func (m *Model) render() string {
	if m.width <= 0 || m.height <= 0 {
		return "loading..."
	}

	header := m.renderHeader()
	main := m.renderMain()
	footer := m.renderFooter()
	return lipgloss.JoinVertical(lipgloss.Left, header, main, footer)
}

func (m *Model) renderHeader() string {
	focus := "markdown"
	if m.focus == PaneOutline {
		focus = "outline"
	}
	text := fmt.Sprintf("rundown  %s  focus:%s  line:%d/%d", m.fileName, focus, m.cursorLine+1, max(1, len(m.doc.Lines)))
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("24")).
		Padding(0, 1).
		Width(m.width)
	return style.Render(text)
}

func (m *Model) renderMain() string {
	mainH := max(4, m.mainHeight())
	leftW := max(20, (m.width*2)/3)
	rightW := max(20, m.width-leftW)
	leftW = max(20, m.width-rightW)

	left := m.renderMarkdownPane(leftW, mainH)
	right := m.renderOutlinePane(rightW, mainH)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m *Model) renderFooter() string {
	text := "tab pane  |  quit: C-c C-q Q  |  markdown: hjkl HJKL  |  outline: j/k c/C e/E x n p r"
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Width(m.width)
	return style.Render(text)
}

func (m *Model) renderMarkdownPane(width, height int) string {
	active := m.focus == PaneMarkdown
	bodyW := max(8, width-4)
	bodyH := max(1, height-2)

	markdown := m.renderMarkdown(bodyH, bodyW)
	body := lipgloss.NewStyle().
		Width(bodyW).
		Height(bodyH).
		MaxHeight(bodyH).
		MaxWidth(bodyW).
		Render(markdown)

	borderColor := lipgloss.Color("240")
	if active {
		borderColor = lipgloss.Color("45")
	}
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Render(body)
}

func (m *Model) renderMarkdown(height, width int) string {
	if len(m.doc.Lines) == 0 || height <= 0 {
		return ""
	}

	selStart, selEnd, hasSelection := m.selectedLineRange()
	lines := make([]string, 0, height)

	for row := 0; row < height; row++ {
		lineIdx := m.mdTop + row
		lines = append(lines, m.renderMarkdownRow(lineIdx, width, selStart, selEnd, hasSelection))
	}

	return strings.Join(lines, "\n")
}

func (m *Model) renderMarkdownRow(lineIdx, width, selStart, selEnd int, hasSelection bool) string {
	if lineIdx >= len(m.doc.Lines) {
		return padLine("", width)
	}

	content := renderMarkdownLine(m.doc.Lines[lineIdx])
	selected := hasSelection && lineIdx >= selStart && lineIdx <= selEnd
	prefix := selectionPrefix(lineIdx, selStart, selEnd, selected)
	if lineIdx == m.cursorLine {
		prefix = "▸ "
	}

	style := markdownLineStyle(lineIdx == m.cursorLine, selected)
	return padLine(style.Render(prefix+content), width)
}

func selectionPrefix(lineIdx, selStart, selEnd int, selected bool) string {
	if !selected {
		return "  "
	}
	switch {
	case lineIdx == selStart && lineIdx == selEnd:
		return "◆ "
	case lineIdx == selStart:
		return "┌ "
	case lineIdx == selEnd:
		return "└ "
	default:
		return "│ "
	}
}

func markdownLineStyle(isCursor, selected bool) lipgloss.Style {
	style := lipgloss.NewStyle()
	if selected {
		style = style.Background(lipgloss.Color("236"))
	}
	if isCursor {
		style = style.Background(lipgloss.Color("24")).Foreground(lipgloss.Color("230")).Bold(true)
	}
	return style
}

func renderMarkdownLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "#") {
		level := headingLevel(trimmed)
		if level > 0 {
			title := strings.TrimSpace(trimmed[level:])
			return fmt.Sprintf("%s %s", strings.Repeat("•", level), title)
		}
	}
	return line
}

func (m *Model) renderOutlinePane(width, height int) string {
	active := m.focus == PaneOutline
	bodyW := max(8, width-4)
	bodyH := max(1, height-2)
	lines := m.renderOutlineLines(bodyH, bodyW)
	body := lipgloss.NewStyle().
		Width(bodyW).
		Height(bodyH).
		Render(strings.Join(lines, "\n"))

	borderColor := lipgloss.Color("240")
	if active {
		borderColor = lipgloss.Color("205")
	}
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Render(body)
}

func (m *Model) renderOutlineLines(height, width int) []string {
	visible := m.visibleOutlineIndexes()
	pos := slicesIndex(visible, m.outlineIdx)
	if pos < 0 {
		pos = 0
	}

	top := 0
	if pos >= height {
		top = pos - height + 1
	}

	lines := make([]string, 0, height)
	for row := 0; row < height; row++ {
		idx := top + row
		if idx >= len(visible) {
			lines = append(lines, padLine("", width))
			continue
		}
		itemIdx := visible[idx]
		item := m.doc.Outline[itemIdx]
		lines = append(lines, padLine(m.renderOutlineItem(itemIdx, item), width))
	}
	return lines
}

func (m *Model) renderOutlineItem(idx int, item OutlineItem) string {
	cursor := " "
	if idx == m.outlineIdx {
		cursor = "▸"
	}
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	if idx == m.outlineIdx {
		cursorStyle = cursorStyle.Bold(true)
	}

	if item.Kind == NodeExec {
		line := fmt.Sprintf("%s %s %s", cursorStyle.Render(cursor), iconForLang(item.Lang), item.Title)
		if idx == m.outlineIdx {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(line)
		}
		return line
	}

	collapse := "▾"
	if m.collapsed[idx] {
		collapse = "▸"
	}
	indent := strings.Repeat("  ", max(0, item.Level-1))
	line := fmt.Sprintf("%s %s%s %s", cursorStyle.Render(cursor), indent, collapse, item.Title)
	if idx == m.outlineIdx {
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("87")).Render(line)
	}
	return line
}

func iconForLang(lang string) string {
	switch lang {
	case "bash", "sh", "zsh":
		return ""
	case "go":
		return ""
	case "python", "py":
		return ""
	case "javascript", "js":
		return ""
	case "typescript", "ts":
		return ""
	default:
		return "▶"
	}
}

func fitRightPad(s string, width int) string {
	return padLine(s, width)
}

func slicesIndex(items []int, target int) int {
	for i, item := range items {
		if item == target {
			return i
		}
	}
	return -1
}

func padLine(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) > width {
		return lipgloss.NewStyle().MaxWidth(width).Render(s)
	}
	return s + strings.Repeat(" ", width-lipgloss.Width(s))
}
