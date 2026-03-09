package tui

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	lipgloss "charm.land/lipgloss/v2"
)

func (m *Model) render() string {
	if m.width <= 0 || m.height <= 0 {
		return "loading..."
	}
	if m.helpVisible {
		return constrainViewport(m.renderHelpOverlay(), m.width, m.height)
	}

	header := m.renderHeader()
	main := m.renderMain()
	footer := m.renderFooter()
	parts := []string{header, main}
	if m.execPanelVisible {
		parts = append(parts, m.renderLogPanel(m.logPanelHeight()))
	}
	parts = append(parts, footer)
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return constrainViewport(content, m.width, m.height)
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
	mainH := m.mainHeight()
	leftW := max(20, (m.width*2)/3)
	rightW := max(20, m.width-leftW)
	leftW = max(20, m.width-rightW)

	left := m.renderMarkdownPane(leftW, mainH)
	right := m.renderOutlinePane(rightW, mainH)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m *Model) renderFooter() string {
	text := "? help | tab pane | v hide/show panel | [/] prev/next run | PgUp/PgDn scroll log | outline: r run s stop"
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Width(m.width)
	return style.Render(text)
}

func (m *Model) renderHelpOverlay() string {
	panelFg := lipgloss.Color("252")
	border := lipgloss.Color("110")
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Background(lipgloss.Color("31")).Padding(0, 1)
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("222"))
	bodyStyle := lipgloss.NewStyle().Foreground(panelFg)
	noteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("151")).Italic(true)

	lines := []string{
		titleStyle.Render("rundown help"),
		"",
		sectionStyle.Render("Global"),
		bodyStyle.Render("  ?        toggle this help"),
		bodyStyle.Render("  tab      switch pane focus"),
		bodyStyle.Render("  Ctrl+A   jump to document top"),
		bodyStyle.Render("  Ctrl+E   jump to document bottom"),
		bodyStyle.Render("  Ctrl+C/Q quit (stops running commands)"),
		"",
		sectionStyle.Render("Markdown pane"),
		bodyStyle.Render("  j/k      move cursor down/up"),
		bodyStyle.Render("  h/l      fallback left/right navigation"),
		bodyStyle.Render("  J/K      next/previous heading"),
		bodyStyle.Render("  H/L      parent/first child heading"),
		bodyStyle.Render("  mouse    wheel scroll (left pane)"),
		"",
		sectionStyle.Render("Outline pane"),
		bodyStyle.Render("  j/k      move selection"),
		bodyStyle.Render("  c/C      collapse current/all headings"),
		bodyStyle.Render("  e/E      expand current/all headings"),
		bodyStyle.Render("  x        toggle executable-only view"),
		bodyStyle.Render("  n/p      next/previous executable"),
		bodyStyle.Render("  r        run selected executable"),
		bodyStyle.Render("  s        stop running command"),
		"",
		sectionStyle.Render("Execution panel"),
		bodyStyle.Render("  v        show/hide panel"),
		bodyStyle.Render("  [/]      previous/next run for selected block"),
		bodyStyle.Render("  PgUp/Dn  scroll logs"),
		bodyStyle.Render("  Ctrl+U/D scroll logs"),
		bodyStyle.Render("  Home/End jump to top/bottom logs"),
		bodyStyle.Render("  mouse    wheel scroll (over log panel)"),
		"",
		noteStyle.Render("Press ? or Esc to close help."),
	}
	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Foreground(panelFg).
		Render(body)
}

func (m *Model) renderLogPanel(height int) string {
	panelW := max(8, m.width)
	bodyW := max(1, panelW-4) // border + horizontal padding
	rec, current, total, ok := m.execPanelData()
	if !ok {
		rec = ExecRecord{Title: "none", Lang: "", Status: "idle"}
	}
	status := rec.Status
	titleLine := m.renderExecTitleLine(rec, current, total, bodyW)
	metaLine := m.renderExecMetaLine(rec, current, total, bodyW)
	visible := max(1, height-4) // border top/bottom + title + meta
	renderedLogs := m.visibleExecLogs(rec)
	maxScroll := max(0, len(renderedLogs)-visible)
	m.execLogScroll = clamp(m.execLogScroll, 0, maxScroll)
	start := m.execLogScroll
	end := min(len(renderedLogs), start+visible)
	lines := make([]string, 0, visible)
	for _, line := range renderedLogs[start:end] {
		lines = append(lines, m.renderExecLogLine(line, bodyW))
	}
	for len(lines) < visible {
		lines = append(lines, strings.Repeat(" ", bodyW))
	}
	content := strings.Join(lines, "\n")

	panel := lipgloss.NewStyle().
		Width(panelW).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(execBorderColor(status)).
		Padding(0, 1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				titleLine,
				metaLine,
				lipgloss.NewStyle().Width(bodyW).Height(visible).Render(content),
			),
		)
	return panel
}

func (m *Model) visibleExecLogs(rec ExecRecord) []ExecLogLine {
	out := make([]ExecLogLine, 0, len(rec.Logs))
	for _, line := range rec.Logs {
		if line.Kind == "command" {
			continue
		}
		out = append(out, line)
	}
	return out
}

func (m *Model) renderExecTitleLine(rec ExecRecord, current, total, width int) string {
	progStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("24")).
		Padding(0, 1)
	runStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("236")).
		Background(lipgloss.Color("186")).
		Padding(0, 1)
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230"))

	program := execProgramName(rec)
	if program == "" {
		program = rec.Lang
	}
	left := progStyle.Render(strings.ToUpper(program))
	right := ""
	if total > 0 {
		right = runStyle.Render(fmt.Sprintf("run %d/%d", current, total))
	}

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	gap := 1
	centerW := max(0, width-leftW-rightW-gap)
	center := titleStyle.Render(clipLine(rec.Title, centerW))
	line := left
	if centerW > 0 {
		line += " " + center
	}
	if right != "" {
		padding := max(1, width-lipgloss.Width(line)-rightW)
		line += strings.Repeat(" ", padding) + right
	}
	return clipLine(line, width)
}

func (m *Model) renderExecMetaLine(rec ExecRecord, current, total, width int) string {
	cmdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	statusStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("231")).
		Background(execStatusAccent(rec.Status)).
		Padding(0, 1)

	parts := []string{}
	if rec.Command != "" {
		parts = append(parts, cmdStyle.Render("$ "+rec.Command))
	}
	if rec.Status != "" {
		statusText := rec.Status
		if rec.Status == "running" {
			statusText = fmt.Sprintf("running | elapsed %s", time.Since(m.execStartedAt).Truncate(time.Second))
		}
		if rec.Duration > 0 && rec.Status != "running" {
			statusText = fmt.Sprintf("%s | %s", rec.Status, rec.Duration.Truncate(time.Millisecond))
		}
		parts = append(parts, statusStyle.Render(statusText))
	}
	return clipLine(strings.Join(parts, "   "), width)
}

func (m *Model) renderExecLogLine(line ExecLogLine, width int) string {
	text := clipLine(line.Text, width)
	switch {
	case line.Kind == "command":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(text)
	case line.Kind == "result":
		return lipgloss.NewStyle().Foreground(execStatusAccent(text)).Bold(true).Render(text)
	case line.Stream == "stderr":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(text)
	default:
		return text
	}
}

func execStatusAccent(status string) color.Color {
	switch {
	case status == "running":
		return lipgloss.Color("180")
	case strings.HasPrefix(status, "completed"):
		return lipgloss.Color("42")
	case strings.HasPrefix(status, "failed"):
		return lipgloss.Color("196")
	case strings.HasPrefix(status, "killed"):
		return lipgloss.Color("208")
	default:
		return lipgloss.Color("244")
	}
}

func execProgramName(rec ExecRecord) string {
	cmd := strings.TrimSpace(rec.Command)
	if cmd == "" {
		return ""
	}
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func clipLine(s string, width int) string {
	if width <= 0 {
		return ""
	}
	clipped := lipgloss.NewStyle().Inline(true).MaxWidth(width).Render(s)
	w := lipgloss.Width(clipped)
	if w >= width {
		return clipped
	}
	return clipped + strings.Repeat(" ", width-w)
}

func constrainViewport(content string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	for i := range lines {
		lines[i] = clipLine(lines[i], width)
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, strings.Repeat(" ", width))
	}
	return strings.Join(lines, "\n")
}

func execBorderColor(status string) color.Color {
	switch {
	case status == "running":
		return lipgloss.Color("221") // beige/yellow
	case strings.HasPrefix(status, "completed"):
		return lipgloss.Color("42") // green
	case strings.HasPrefix(status, "failed"):
		return lipgloss.Color("196") // red
	case strings.HasPrefix(status, "killed"):
		return lipgloss.Color("208")
	default:
		return lipgloss.Color("240")
	}
}

func (m *Model) tailExecLogs(n int) []string {
	if n <= 0 || len(m.execLogs) == 0 {
		return nil
	}
	if len(m.execLogs) <= n {
		out := make([]string, 0, len(m.execLogs))
		for _, line := range m.execLogs {
			out = append(out, line.Text)
		}
		return out
	}
	out := make([]string, 0, n)
	for _, line := range m.execLogs[len(m.execLogs)-n:] {
		out = append(out, line.Text)
	}
	return out
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
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Render(body)
}

func (m *Model) renderMarkdown(height, width int) string {
	if len(m.doc.Lines) == 0 || height <= 0 {
		return ""
	}

	gutterW := 2
	bodyW := max(1, width-gutterW-1)
	bodyLines := m.renderMarkdownBody(height, bodyW)
	selStart, selEnd, hasSelection := m.selectedLineRange()
	rows := make([]string, 0, height)

	for row := 0; row < height; row++ {
		lineIdx := m.mdTop + row
		bar := m.markdownSelectionBar(lineIdx, selStart, selEnd, hasSelection)
		content := ""
		if row < len(bodyLines) {
			content = bodyLines[row]
		}
		rows = append(rows, padLine(bar+" "+content, width))
	}

	return strings.Join(rows, "\n")
}

func (m *Model) renderMarkdownBody(height, width int) []string {
	start := clamp(m.mdTop, 0, max(0, len(m.doc.Lines)-1))
	if m.mdCacheLines != nil &&
		m.mdCacheStart == start &&
		m.mdCacheHeight == height &&
		m.mdCacheWidth == width &&
		m.mdCacheLineCount == len(m.doc.Lines) {
		return m.mdCacheLines
	}

	lines := m.markdownFallbackLines(start, height, width)
	m.storeMarkdownCache(start, height, width, lines)
	return lines
}

func (m *Model) storeMarkdownCache(start, height, width int, lines []string) {
	m.mdCacheStart = start
	m.mdCacheHeight = height
	m.mdCacheWidth = width
	m.mdCacheLineCount = len(m.doc.Lines)
	m.mdCacheLines = lines
}

func (m *Model) markdownFallbackLines(start, height, width int) []string {
	lines := make([]string, 0, height)
	state := concealStateAt(m.doc.Lines, start)
	for row := 0; row < height; row++ {
		lineIdx := start + row
		if lineIdx >= len(m.doc.Lines) {
			lines = append(lines, "")
			continue
		}
		lines = append(lines, padLine(concealMarkdownLine(m.doc.Lines[lineIdx], &state), width))
	}
	return lines
}

func (m *Model) markdownSelectionBar(lineIdx, selStart, selEnd int, hasSelection bool) string {
	selected := hasSelection && lineIdx >= selStart && lineIdx <= selEnd
	prefix := selectionPrefix(lineIdx, selStart, selEnd, selected)
	if lineIdx == m.cursorLine {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("87")).Bold(true).Render("▌ ")
	}
	if !selected {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("239")).Render("│ ")
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render(prefix)
}

func selectionPrefix(lineIdx, selStart, selEnd int, selected bool) string {
	if !selected {
		return "│ "
	}
	switch {
	case lineIdx == selStart && lineIdx == selEnd:
		return "▣ "
	case lineIdx == selStart:
		return "┏ "
	case lineIdx == selEnd:
		return "┗ "
	default:
		return "┃ "
	}
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
		meta := m.execMetaBadge(idx)
		indent := m.execIndent(item)
		line := fmt.Sprintf("%s %s%s %s %s", cursorStyle.Render(cursor), indent, iconForLang(item.Lang), item.Title, meta)
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
	case "ruby", "rb":
		return ""
	case "go":
		return ""
	case "python", "py":
		return ""
	case "javascript", "js":
		return ""
	case "typescript", "ts":
		return ""
	case "rust", "rs":
		return ""
	case "java":
		return ""
	case "kotlin", "kt":
		return ""
	case "php":
		return ""
	case "lua":
		return ""
	case "json":
		return ""
	case "yaml", "yml":
		return ""
	case "dockerfile":
		return ""
	default:
		return "▶"
	}
}

func (m *Model) execMetaBadge(idx int) string {
	if !m.validOutlineIndex(idx) {
		return ""
	}
	blockID := m.doc.Outline[idx].ID
	h := m.execHistory[blockID]
	count := len(h)
	runs := lipgloss.NewStyle().Foreground(lipgloss.Color("110")).Render(fmt.Sprintf("runs:%d", count))
	if count == 0 {
		return runs
	}
	last := h[len(h)-1]

	statusColor := lipgloss.Color("244")
	switch {
	case m.execRunning && m.execRunBlockID == blockID:
		statusColor = lipgloss.Color("214")
	case strings.HasPrefix(last.Status, "completed"):
		statusColor = lipgloss.Color("42")
	case strings.HasPrefix(last.Status, "failed"):
		statusColor = lipgloss.Color("196")
	case strings.HasPrefix(last.Status, "killed"):
		statusColor = lipgloss.Color("208")
	}
	status := lipgloss.NewStyle().Foreground(statusColor).Render(last.Status)

	if m.execRunning && m.execRunBlockID == blockID {
		return fmt.Sprintf("%s | %s", runs, status)
	}
	if last.Duration > 0 {
		d := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(last.Duration.Truncate(time.Millisecond).String())
		return fmt.Sprintf("%s | %s | %s", runs, status, d)
	}
	return fmt.Sprintf("%s | %s", runs, status)
}

func (m *Model) execIndent(item OutlineItem) string {
	if m.execOnly {
		return ""
	}
	if item.Parent >= 0 && item.Parent < len(m.doc.Outline) {
		parentLevel := m.doc.Outline[item.Parent].Level
		return strings.Repeat("  ", max(0, parentLevel))
	}
	return ""
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
