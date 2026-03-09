package tui

import (
	"bufio"
	"fmt"
	"strings"
)

type NodeKind int

const (
	NodeHeading NodeKind = iota
	NodeExec
)

type OutlineItem struct {
	Title   string
	Level   int
	Line    int
	EndLine int
	Kind    NodeKind
	Lang    string
	Parent  int
}

type Document struct {
	Lines   []string
	Outline []OutlineItem
}

// ParseMarkdown builds a lightweight document model for rendering and navigation.
func ParseMarkdown(src string) Document {
	lines := make([]string, 0, 256)
	outline := make([]OutlineItem, 0)
	headingAtLevel := map[int]int{}
	inFence := false
	openExecOutlineIdx := -1

	scanner := bufio.NewScanner(strings.NewReader(src))
	// Allow larger markdown lines/code lines without scanner truncation.
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 4*1024*1024)

	lineNo := 0
	for scanner.Scan() {
		raw := strings.TrimSuffix(scanner.Text(), "\r")
		lines = append(lines, raw)
		trimmed := strings.TrimSpace(raw)

		if !inFence && strings.HasPrefix(trimmed, "#") {
			level := headingLevel(trimmed)
			if level > 0 {
				title := strings.TrimSpace(trimmed[level:])
				parent := nearestHeadingParent(headingAtLevel, level)
				outline = append(outline, OutlineItem{
					Title:   title,
					Level:   level,
					Line:    lineNo,
					EndLine: lineNo,
					Kind:    NodeHeading,
					Parent:  parent,
				})
				headingAtLevel[level] = len(outline) - 1
				for l := level + 1; l <= 6; l++ {
					delete(headingAtLevel, l)
				}
				lineNo++
				continue
			}
		}

		if strings.HasPrefix(trimmed, "```") {
			if inFence {
				inFence = false
				if openExecOutlineIdx >= 0 && openExecOutlineIdx < len(outline) {
					outline[openExecOutlineIdx].EndLine = lineNo
				}
				openExecOutlineIdx = -1
				lineNo++
				continue
			}

			inFence = true
			lang := strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
			if lang == "" {
				lang = "text"
			}
			parent := nearestHeadingParent(headingAtLevel, 7)
			outline = append(outline, OutlineItem{
				Title:   fmt.Sprintf("%s block", lang),
				Level:   7,
				Line:    lineNo,
				EndLine: lineNo,
				Kind:    NodeExec,
				Lang:    strings.ToLower(lang),
				Parent:  parent,
			})
			openExecOutlineIdx = len(outline) - 1
		}
		lineNo++
	}

	// Match strings.Split behavior that keeps a trailing empty line.
	if strings.HasSuffix(src, "\n") {
		lines = append(lines, "")
	}
	lastLine := docMax(0, len(lines)-1)
	if openExecOutlineIdx >= 0 && openExecOutlineIdx < len(outline) {
		outline[openExecOutlineIdx].EndLine = lastLine
	}
	assignHeadingEndLines(outline, lastLine)

	return Document{
		Lines:   lines,
		Outline: outline,
	}
}

func assignHeadingEndLines(outline []OutlineItem, lastLine int) {
	nextHeadingAtLevel := [7]int{}
	for i := range nextHeadingAtLevel {
		nextHeadingAtLevel[i] = -1
	}

	for i := len(outline) - 1; i >= 0; i-- {
		if outline[i].Kind != NodeHeading {
			if outline[i].EndLine == 0 {
				outline[i].EndLine = lastLine
			}
			continue
		}

		end := lastLine
		for lvl := 1; lvl <= outline[i].Level; lvl++ {
			if next := nextHeadingAtLevel[lvl]; next >= 0 {
				if next-1 < end {
					end = next - 1
				}
			}
		}
		if end < outline[i].Line {
			end = outline[i].Line
		}
		outline[i].EndLine = end
		nextHeadingAtLevel[outline[i].Level] = outline[i].Line
	}
}

func docMax(a, b int) int {
	if a > b {
		return a
	}
	return b
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
