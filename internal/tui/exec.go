package tui

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

type execLineMsg struct {
	line   string
	stream string
}

type ExecLogLine struct {
	Text   string
	Stream string
	Kind   string
}

type ExecRecord struct {
	Title    string
	Lang     string
	Command  string
	Started  time.Time
	Duration time.Duration
	Status   string
	ExitCode int
	Logs     []ExecLogLine
}

type execDoneMsg struct {
	err      error
	exitCode int
	killed   bool
	duration time.Duration
}

type execChannelClosedMsg struct{}

func (m *Model) runSelectedExecutable() tea.Cmd {
	item, script, ok := m.selectedExecutableBlock()
	if !ok {
		return nil
	}
	if m.execRunning {
		return nil
	}

	name, args := commandForLanguage(item.Lang, script)
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, name, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return func() tea.Msg { return execDoneMsg{err: err, exitCode: -1} }
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return func() tea.Msg { return execDoneMsg{err: err, exitCode: -1} }
	}
	if err := cmd.Start(); err != nil {
		return func() tea.Msg { return execDoneMsg{err: err, exitCode: -1} }
	}

	m.execPanelVisible = true
	m.execRunning = true
	m.execCancel = cancel
	m.execStartedAt = time.Now()
	m.execTitle = fmt.Sprintf("%s (%s)", item.Title, item.Lang)
	cmdLine := fmt.Sprintf("$ %s %s", name, strings.Join(args, " "))
	m.execLogs = []ExecLogLine{{Text: cmdLine, Kind: "command"}}
	m.execStatus = "running"
	m.execMsgCh = make(chan tea.Msg, 1024)
	m.execRunBlockID = item.ID

	record := ExecRecord{
		Title:   item.Title,
		Lang:    item.Lang,
		Command: fmt.Sprintf("%s %s", name, strings.Join(args, " ")),
		Started: m.execStartedAt,
		Status:  "running",
		Logs:    append([]ExecLogLine{}, m.execLogs...),
	}
	m.execHistory[m.execRunBlockID] = append(m.execHistory[m.execRunBlockID], record)
	m.execViewIndex[m.execRunBlockID] = len(m.execHistory[m.execRunBlockID]) - 1
	m.execViewBlockID = m.execRunBlockID
	m.execFollowTail = true
	m.execLogScroll = 0

	go streamExecPipe(stdout, "stdout", m.execMsgCh)
	go streamExecPipe(stderr, "stderr", m.execMsgCh)
	go waitExec(cmd, ctx, m.execStartedAt, m.execMsgCh)

	return waitExecEvent(m.execMsgCh)
}

func (m *Model) selectedExecutableBlock() (OutlineItem, string, bool) {
	if !m.validOutlineIndex(m.outlineIdx) {
		return OutlineItem{}, "", false
	}
	item := m.doc.Outline[m.outlineIdx]
	if item.Kind != NodeExec {
		return OutlineItem{}, "", false
	}

	start := item.Line + 1
	end := item.EndLine
	if start < 0 || start >= len(m.doc.Lines) {
		return OutlineItem{}, "", false
	}
	if end < start {
		end = start
	}
	if end > len(m.doc.Lines) {
		end = len(m.doc.Lines)
	}

	script := strings.Join(m.doc.Lines[start:end], "\n")
	return item, script, true
}

func (m *Model) stopExecution() {
	if m.execRunning && m.execCancel != nil {
		m.execCancel()
	}
}

func waitExecEvent(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return execChannelClosedMsg{}
		}
		return msg
	}
}

func streamExecPipe(r io.Reader, stream string, out chan<- tea.Msg) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 2*1024*1024)
	for scanner.Scan() {
		out <- execLineMsg{line: scanner.Text(), stream: stream}
	}
}

func waitExec(cmd *exec.Cmd, ctx context.Context, started time.Time, out chan<- tea.Msg) {
	err := cmd.Wait()
	killed := errors.Is(ctx.Err(), context.Canceled)
	out <- execDoneMsg{
		err:      err,
		exitCode: exitCode(err),
		killed:   killed,
		duration: time.Since(started),
	}
	close(out)
}

func commandForLanguage(lang, script string) (string, []string) {
	switch strings.ToLower(lang) {
	case "bash", "sh", "zsh", "shell":
		return "bash", []string{"-lc", script}
	case "ruby", "rb":
		return "ruby", []string{"-e", script}
	case "python", "py":
		return "python3", []string{"-c", script}
	case "javascript", "js":
		return "node", []string{"-e", script}
	default:
		return "bash", []string{"-lc", script}
	}
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}
