package tui

import "fmt"

func (m *Model) toggleExecPanel() {
	m.execPanelVisible = !m.execPanelVisible
}

func (m *Model) panelOutline() int {
	if m.execRunning {
		return m.outlineIndexByID(m.execRunBlockID)
	}
	if m.validOutlineIndex(m.outlineIdx) && m.doc.Outline[m.outlineIdx].Kind == NodeExec {
		return m.outlineIdx
	}
	return m.outlineIndexByID(m.execViewBlockID)
}

func (m *Model) ensureExecViewIndex(blockID string) {
	if blockID == "" {
		return
	}
	history := m.execHistory[blockID]
	if len(history) == 0 {
		m.execViewIndex[blockID] = 0
		return
	}
	cur, ok := m.execViewIndex[blockID]
	if !ok || cur < 0 || cur >= len(history) {
		m.execViewIndex[blockID] = len(history) - 1
	}
}

func (m *Model) execPrevRecord() {
	outline := m.panelOutline()
	if outline < 0 {
		return
	}
	blockID := m.doc.Outline[outline].ID
	m.ensureExecViewIndex(blockID)
	if m.execViewIndex[blockID] > 0 {
		m.execViewIndex[blockID]--
		m.execLogScroll = 0
		m.execViewBlockID = blockID
	}
}

func (m *Model) execNextRecord() {
	outline := m.panelOutline()
	if outline < 0 {
		return
	}
	blockID := m.doc.Outline[outline].ID
	m.ensureExecViewIndex(blockID)
	history := m.execHistory[blockID]
	if len(history) == 0 {
		return
	}
	if m.execViewIndex[blockID] < len(history)-1 {
		m.execViewIndex[blockID]++
		m.execLogScroll = 0
		m.execViewBlockID = blockID
	}
}

func (m *Model) execCount(idx int) int {
	if !m.validOutlineIndex(idx) {
		return 0
	}
	return len(m.execHistory[m.doc.Outline[idx].ID])
}

func (m *Model) execScroll(delta int) {
	outline := m.panelOutline()
	if outline < 0 {
		return
	}
	_, logs, _, _, _ := m.execPanelData()
	maxScroll := max(0, len(logs)-max(1, m.logPanelHeight()-3))
	m.execLogScroll = clamp(m.execLogScroll+delta, 0, maxScroll)
}

func (m *Model) execPanelData() (string, []string, string, int, int) {
	outline := m.panelOutline()
	if outline < 0 {
		return "exec: none", nil, "idle", 0, 0
	}
	blockID := m.doc.Outline[outline].ID
	m.ensureExecViewIndex(blockID)
	history := m.execHistory[blockID]
	if len(history) == 0 {
		return "exec: none", nil, "idle", 0, 0
	}
	idx := m.execViewIndex[blockID]
	rec := history[idx]
	title := fmt.Sprintf("exec: %s (%s)", rec.Title, rec.Lang)
	return title, rec.Logs, rec.Status, idx + 1, len(history)
}

func (m *Model) outlineIndexByID(blockID string) int {
	if blockID == "" {
		return -1
	}
	for i, item := range m.doc.Outline {
		if item.ID == blockID {
			return i
		}
	}
	return -1
}
