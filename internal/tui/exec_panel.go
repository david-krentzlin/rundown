package tui

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
		m.execFollowTail = false
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
		m.execFollowTail = false
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
	if delta < 0 {
		m.execFollowTail = false
	}
	rec, _, _, ok := m.execPanelData()
	if !ok {
		return
	}
	maxScroll := max(0, len(m.visibleExecLogs(rec))-max(1, m.logPanelHeight()-4))
	m.execLogScroll = clamp(m.execLogScroll+delta, 0, maxScroll)
	if m.execLogScroll >= maxScroll {
		m.execFollowTail = true
	}
}

func (m *Model) execPanelData() (ExecRecord, int, int, bool) {
	outline := m.panelOutline()
	if outline < 0 {
		return ExecRecord{}, 0, 0, false
	}
	blockID := m.doc.Outline[outline].ID
	m.ensureExecViewIndex(blockID)
	history := m.execHistory[blockID]
	if len(history) == 0 {
		return ExecRecord{}, 0, 0, false
	}
	idx := m.execViewIndex[blockID]
	return history[idx], idx + 1, len(history), true
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

func (m *Model) isViewingActiveRun() bool {
	outline := m.panelOutline()
	if outline < 0 || !m.validOutlineIndex(outline) {
		return false
	}
	return m.doc.Outline[outline].ID == m.execRunBlockID
}
