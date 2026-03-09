package tui

import (
	"fmt"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
)

type concealState struct {
	inFence bool
}

func concealStateAt(lines []string, start int) concealState {
	st := concealState{}
	if start <= 0 {
		return st
	}
	limit := min(start, len(lines))
	for i := 0; i < limit; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "```") {
			st.inFence = !st.inFence
		}
	}
	return st
}

func concealMarkdownLine(line string, st *concealState) string {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "```") {
		st.inFence = !st.inFence
		return lipgloss.NewStyle().Foreground(lipgloss.Color("109")).Render(line)
	}
	if st.inFence {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(line)
	}
	if h := headingLevel(trimmed); h > 0 {
		title := strings.TrimSpace(trimmed[h:])
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")).Render(fmt.Sprintf("%s %s", strings.Repeat("•", h), title))
	}
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("151")).Render(line)
	}
	if strings.HasPrefix(trimmed, ">") {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("110")).Render(line)
	}
	return line
}
