package tui

import "fmt"

func (m *Model) toggleExecPanel() {
	m.execPanelVisible = !m.execPanelVisible
}

func (m *Model) panelOutline() int {
	if m.execRunning {
		return m.execRunOutlineIdx
	}
	if m.validOutlineIndex(m.outlineIdx) && m.doc.Outline[m.outlineIdx].Kind == NodeExec {
		return m.outlineIdx
	}
	return m.execViewOutline
}

func (m *Model) ensureExecViewIndex(outline int) {
	if outline < 0 {
		return
	}
	history := m.execHistory[outline]
	if len(history) == 0 {
		m.execViewIndex[outline] = 0
		return
	}
	cur, ok := m.execViewIndex[outline]
	if !ok || cur < 0 || cur >= len(history) {
		m.execViewIndex[outline] = len(history) - 1
	}
}

func (m *Model) execPrevRecord() {
	outline := m.panelOutline()
	if outline < 0 {
		return
	}
	m.ensureExecViewIndex(outline)
	if m.execViewIndex[outline] > 0 {
		m.execViewIndex[outline]--
		m.execLogScroll = 0
		m.execViewOutline = outline
	}
}

func (m *Model) execNextRecord() {
	outline := m.panelOutline()
	if outline < 0 {
		return
	}
	m.ensureExecViewIndex(outline)
	history := m.execHistory[outline]
	if len(history) == 0 {
		return
	}
	if m.execViewIndex[outline] < len(history)-1 {
		m.execViewIndex[outline]++
		m.execLogScroll = 0
		m.execViewOutline = outline
	}
}

func (m *Model) execCount(idx int) int {
	return len(m.execHistory[idx])
}

func (m *Model) execScroll(delta int) {
	outline := m.panelOutline()
	if outline < 0 {
		return
	}
	_, logs, _, _, _ := m.execPanelData()
	maxScroll := max(0, len(logs)-max(1, m.logPanelHeight()-4))
	m.execLogScroll = clamp(m.execLogScroll+delta, 0, maxScroll)
}

func (m *Model) execPanelData() (string, []string, string, int, int) {
	outline := m.panelOutline()
	if outline < 0 {
		return "exec: none", nil, "idle", 0, 0
	}
	m.ensureExecViewIndex(outline)
	history := m.execHistory[outline]
	if len(history) == 0 {
		return "exec: none", nil, "idle", 0, 0
	}
	idx := m.execViewIndex[outline]
	rec := history[idx]
	title := fmt.Sprintf("exec: %s (%s)", rec.Title, rec.Lang)
	return title, rec.Logs, rec.Status, idx + 1, len(history)
}
