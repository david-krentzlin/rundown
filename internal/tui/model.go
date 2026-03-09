package tui

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type Pane int

const (
	PaneMarkdown Pane = iota
	PaneOutline
)

type Model struct {
	doc      Document
	fileName string
	focus    Pane

	width      int
	height     int
	cursorLine int
	mdTop      int
	outlineIdx int

	collapsed map[int]bool
	execOnly  bool

	mdCacheStart     int
	mdCacheHeight    int
	mdCacheWidth     int
	mdCacheLineCount int
	mdCacheLines     []string
}

func NewModel(doc Document, fileName string) *Model {
	return &Model{
		doc:       doc,
		fileName:  fileName,
		focus:     PaneMarkdown,
		collapsed: map[int]bool{},
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch x := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetViewport(x.Width, x.Height)
		return m, nil
	case tea.KeyMsg:
		if _, isRelease := x.(tea.KeyReleaseMsg); isRelease {
			return m, nil
		}
		if m.handleKey(x.String()) {
			return m, tea.Quit
		}
	case tea.MouseWheelMsg:
		m.handleMouseWheel(x)
		return m, nil
	}

	return m, nil
}

func (m *Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (m *Model) SetViewport(width, height int) {
	m.width = max(width, 20)
	m.height = max(height, 8)
	m.ensureBounds()
}

func (m *Model) handleKey(key string) bool {
	switch key {
	case "ctrl+c", "ctrl+q", "Q":
		return true
	case "tab", "ctrl+i":
		if m.focus == PaneMarkdown {
			m.focus = PaneOutline
		} else {
			m.focus = PaneMarkdown
		}
		return false
	}

	if m.focus == PaneMarkdown {
		m.handleMarkdownKey(key)
	} else {
		m.handleOutlineKey(key)
	}
	return false
}

func (m *Model) handleMarkdownKey(key string) {
	switch key {
	case "j", "down", "l", "right":
		m.moveMarkdown(1)
	case "k", "up", "h", "left":
		m.moveMarkdown(-1)
	case "J":
		m.jumpHeadingByOrder(1)
	case "K":
		m.jumpHeadingByOrder(-1)
	case "H":
		m.jumpToParentHeading()
	case "L":
		m.jumpToChildHeading()
	}
}

func (m *Model) handleOutlineKey(key string) {
	switch key {
	case "j", "down":
		m.moveOutline(1)
	case "k", "up":
		m.moveOutline(-1)
	case "c":
		m.collapseCurrent()
	case "C":
		m.collapseAll()
	case "e":
		m.expandCurrent()
	case "E":
		m.expandAll()
	case "x":
		m.execOnly = !m.execOnly
		m.ensureOutlineIndexVisible()
	case "n":
		m.jumpExec(1)
	case "p":
		m.jumpExec(-1)
	case "r":
		// reserved for execution in a later iteration
	}
	m.syncMarkdownFromOutline()
}

func (m *Model) handleMouseWheel(msg tea.MouseWheelMsg) {
	mouse := msg.Mouse()
	if !m.isInMarkdownPane(mouse.X, mouse.Y) {
		return
	}

	switch mouse.Button {
	case tea.MouseWheelUp:
		m.scrollMarkdown(-3)
	case tea.MouseWheelDown:
		m.scrollMarkdown(3)
	}
}

func (m *Model) moveMarkdown(delta int) {
	m.cursorLine = clamp(m.cursorLine+delta, 0, max(0, len(m.doc.Lines)-1))
	m.ensureMarkdownCursorVisible()
	m.syncOutlineFromMarkdown()
}

func (m *Model) scrollMarkdown(delta int) {
	maxTop := max(0, len(m.doc.Lines)-m.mainHeight())
	m.mdTop = clamp(m.mdTop+delta, 0, maxTop)
	m.cursorLine = clamp(m.mdTop, 0, max(0, len(m.doc.Lines)-1))
	m.syncOutlineFromMarkdown()
}

func (m *Model) jumpHeadingByOrder(direction int) {
	headings := m.headingIndexes()
	if len(headings) == 0 {
		return
	}

	current := -1
	for _, idx := range headings {
		if m.doc.Outline[idx].Line <= m.cursorLine {
			current = idx
		} else {
			break
		}
	}

	pos := slices.Index(headings, current)
	if pos < 0 {
		if direction > 0 {
			pos = 0
		} else {
			pos = len(headings) - 1
		}
	} else {
		pos += direction
	}
	pos = clamp(pos, 0, len(headings)-1)
	m.outlineIdx = headings[pos]
	m.syncMarkdownFromOutline()
}

func (m *Model) jumpToParentHeading() {
	current := m.currentHeadingIndex()
	if current < 0 {
		return
	}
	parent := m.doc.Outline[current].Parent
	if parent < 0 {
		return
	}
	m.outlineIdx = parent
	m.syncMarkdownFromOutline()
}

func (m *Model) jumpToChildHeading() {
	current := m.currentHeadingIndex()
	if current < 0 {
		return
	}
	for i := current + 1; i < len(m.doc.Outline); i++ {
		item := m.doc.Outline[i]
		if item.Kind != NodeHeading {
			continue
		}
		if item.Parent == current {
			m.outlineIdx = i
			m.syncMarkdownFromOutline()
			return
		}
	}
}

func (m *Model) moveOutline(delta int) {
	visible := m.visibleOutlineIndexes()
	if len(visible) == 0 {
		return
	}
	m.ensureOutlineIndexVisible()
	pos := slices.Index(visible, m.outlineIdx)
	if pos < 0 {
		pos = 0
	}
	pos = clamp(pos+delta, 0, len(visible)-1)
	m.outlineIdx = visible[pos]
}

func (m *Model) collapseCurrent() {
	if !m.validOutlineIndex(m.outlineIdx) {
		return
	}
	if m.doc.Outline[m.outlineIdx].Kind == NodeHeading {
		m.collapsed[m.outlineIdx] = true
	}
}

func (m *Model) collapseAll() {
	for i, item := range m.doc.Outline {
		if item.Kind == NodeHeading {
			m.collapsed[i] = true
		}
	}
	m.ensureOutlineIndexVisible()
}

func (m *Model) expandCurrent() {
	if !m.validOutlineIndex(m.outlineIdx) {
		return
	}
	delete(m.collapsed, m.outlineIdx)
}

func (m *Model) expandAll() {
	clear(m.collapsed)
}

func (m *Model) jumpExec(direction int) {
	execs := make([]int, 0)
	for i, item := range m.doc.Outline {
		if item.Kind == NodeExec {
			execs = append(execs, i)
		}
	}
	if len(execs) == 0 {
		return
	}

	curPos := slices.Index(execs, m.outlineIdx)
	if curPos < 0 {
		curPos = 0
	} else {
		curPos += direction
	}
	curPos = clamp(curPos, 0, len(execs)-1)
	m.outlineIdx = execs[curPos]
}

func (m *Model) syncOutlineFromMarkdown() {
	if len(m.doc.Outline) == 0 {
		return
	}
	candidate := 0
	for i, item := range m.doc.Outline {
		if item.Line <= m.cursorLine {
			candidate = i
		} else {
			break
		}
	}
	m.outlineIdx = candidate
	m.ensureOutlineIndexVisible()
}

func (m *Model) syncMarkdownFromOutline() {
	if !m.validOutlineIndex(m.outlineIdx) {
		return
	}
	m.cursorLine = m.doc.Outline[m.outlineIdx].Line
	m.ensureMarkdownCursorVisible()
}

func (m *Model) ensureMarkdownCursorVisible() {
	mainH := m.mainHeight()
	if m.cursorLine < m.mdTop {
		m.mdTop = m.cursorLine
	}
	if m.cursorLine >= m.mdTop+mainH {
		m.mdTop = m.cursorLine - mainH + 1
	}
	m.ensureBounds()
}

func (m *Model) ensureOutlineIndexVisible() {
	visible := m.visibleOutlineIndexes()
	if len(visible) == 0 {
		m.outlineIdx = 0
		return
	}
	if !slices.Contains(visible, m.outlineIdx) {
		m.outlineIdx = visible[0]
	}
}

func (m *Model) headingIndexes() []int {
	items := make([]int, 0)
	for i, item := range m.doc.Outline {
		if item.Kind == NodeHeading {
			items = append(items, i)
		}
	}
	return items
}

func (m *Model) currentHeadingIndex() int {
	idx := -1
	for i, item := range m.doc.Outline {
		if item.Line > m.cursorLine {
			break
		}
		if item.Kind == NodeHeading {
			idx = i
		}
	}
	if idx >= 0 {
		return idx
	}
	for i, item := range m.doc.Outline {
		if item.Kind == NodeHeading {
			return i
		}
	}
	return -1
}

func (m *Model) visibleOutlineIndexes() []int {
	items := make([]int, 0, len(m.doc.Outline))
	for i, item := range m.doc.Outline {
		if m.execOnly && item.Kind != NodeExec {
			continue
		}
		if m.isHiddenByCollapse(i) {
			continue
		}
		items = append(items, i)
	}
	return items
}

func (m *Model) isHiddenByCollapse(idx int) bool {
	for parent := m.doc.Outline[idx].Parent; parent >= 0; parent = m.doc.Outline[parent].Parent {
		if m.collapsed[parent] {
			return true
		}
	}
	return false
}

func (m *Model) validOutlineIndex(idx int) bool {
	return idx >= 0 && idx < len(m.doc.Outline)
}

func (m *Model) ensureBounds() {
	maxLine := max(0, len(m.doc.Lines)-1)
	m.cursorLine = clamp(m.cursorLine, 0, maxLine)
	maxTop := max(0, len(m.doc.Lines)-m.mainHeight())
	m.mdTop = clamp(m.mdTop, 0, maxTop)
	if len(m.doc.Outline) == 0 {
		m.outlineIdx = 0
		return
	}
	m.outlineIdx = clamp(m.outlineIdx, 0, len(m.doc.Outline)-1)
}

func (m *Model) mainHeight() int {
	return max(3, m.height-2)
}

func (m *Model) leftPaneWidth() int {
	leftW := max(20, (m.width*2)/3)
	rightW := max(20, m.width-leftW)
	leftW = max(20, m.width-rightW)
	return leftW
}

func (m *Model) isInMarkdownPane(x, y int) bool {
	mainYStart := 1
	mainYEnd := mainYStart + m.mainHeight() - 1
	if y < mainYStart || y > mainYEnd {
		return false
	}
	return x >= 0 && x < m.leftPaneWidth()
}

func (m *Model) selectedLineRange() (int, int, bool) {
	if !m.validOutlineIndex(m.outlineIdx) || len(m.doc.Lines) == 0 {
		return 0, 0, false
	}

	item := m.doc.Outline[m.outlineIdx]
	start := clamp(item.Line, 0, len(m.doc.Lines)-1)
	end := len(m.doc.Lines) - 1

	switch item.Kind {
	case NodeHeading:
		for i := m.outlineIdx + 1; i < len(m.doc.Outline); i++ {
			next := m.doc.Outline[i]
			if next.Kind == NodeHeading && next.Level <= item.Level {
				end = next.Line - 1
				break
			}
		}
	case NodeExec:
		// Prefer fence close line when present.
		for line := start + 1; line < len(m.doc.Lines); line++ {
			if strings.HasPrefix(strings.TrimSpace(m.doc.Lines[line]), "```") {
				end = line
				break
			}
		}
	}

	if end < start {
		end = start
	}
	return start, clamp(end, start, len(m.doc.Lines)-1), true
}

func clamp(v, minV, maxV int) int {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}
