package tui

import (
	"fmt"
	"strings"
)

type NodeKind int

const (
	NodeHeading NodeKind = iota
	NodeExec
)

type OutlineItem struct {
	Title  string
	Level  int
	Line   int
	Kind   NodeKind
	Lang   string
	Parent int
}

type Document struct {
	Lines   []string
	Outline []OutlineItem
}

// ParseMarkdown builds a lightweight document model for rendering and navigation.
func ParseMarkdown(src string) Document {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	outline := make([]OutlineItem, 0)
	headingAtLevel := map[int]int{}
	inFence := false

	for i, raw := range lines {
		trimmed := strings.TrimSpace(raw)
		if !inFence && strings.HasPrefix(trimmed, "#") {
			level := headingLevel(trimmed)
			if level > 0 {
				title := strings.TrimSpace(trimmed[level:])
				parent := nearestHeadingParent(headingAtLevel, level)
				outline = append(outline, OutlineItem{
					Title:  title,
					Level:  level,
					Line:   i,
					Kind:   NodeHeading,
					Parent: parent,
				})
				headingAtLevel[level] = len(outline) - 1
				for l := level + 1; l <= 6; l++ {
					delete(headingAtLevel, l)
				}
				continue
			}
		}

		if strings.HasPrefix(trimmed, "```") {
			if inFence {
				inFence = false
				continue
			}

			inFence = true
			lang := strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
			if lang == "" {
				lang = "text"
			}
			parent := nearestHeadingParent(headingAtLevel, 7)
			outline = append(outline, OutlineItem{
				Title:  fmt.Sprintf("%s block", lang),
				Level:  7,
				Line:   i,
				Kind:   NodeExec,
				Lang:   strings.ToLower(lang),
				Parent: parent,
			})
		}
	}

	return Document{
		Lines:   lines,
		Outline: outline,
	}
}

func headingLevel(line string) int {
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	if level == 0 || level > 6 {
		return 0
	}
	if level >= len(line) || line[level] != ' ' {
		return 0
	}
	return level
}

func nearestHeadingParent(levelIdx map[int]int, level int) int {
	for l := level - 1; l >= 1; l-- {
		if idx, ok := levelIdx[l]; ok {
			return idx
		}
	}
	return -1
}
