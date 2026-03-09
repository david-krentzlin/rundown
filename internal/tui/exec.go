package tui

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
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

type shellSession struct {
	key       string
	lang      string
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	currentCh chan tea.Msg
	token     string
	startedAt time.Time
	mu        sync.Mutex
}

func (m *Model) runSelectedExecutable() tea.Cmd {
	item, script, ok := m.selectedExecutableBlock()
	if !ok {
		return nil
	}
	if m.execRunning {
		return nil
	}
	if item.Session && supportsSession(item.Lang) {
		return m.runSelectedSessionExecutable(item, script)
	}

	name, args := commandForLanguage(item.Lang, script, false)
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

func (m *Model) runSelectedSessionExecutable(item OutlineItem, script string) tea.Cmd {
	session, err := m.getOrCreateShellSession(sessionKey(item), item.Lang)
	if err != nil {
		return func() tea.Msg { return execDoneMsg{err: err, exitCode: -1} }
	}

	m.execPanelVisible = true
	m.execRunning = true
	m.execStartedAt = time.Now()
	m.execTitle = fmt.Sprintf("%s (%s)", item.Title, item.Lang)
	m.execLogs = []ExecLogLine{{Text: fmt.Sprintf("$ %s [session]", session.lang), Kind: "command"}}
	m.execStatus = "running"
	m.execMsgCh = make(chan tea.Msg, 1024)
	m.execRunBlockID = item.ID

	record := ExecRecord{
		Title:   item.Title,
		Lang:    item.Lang,
		Command: fmt.Sprintf("%s [session]", session.lang),
		Started: m.execStartedAt,
		Status:  "running",
		Logs:    append([]ExecLogLine{}, m.execLogs...),
	}
	m.execHistory[m.execRunBlockID] = append(m.execHistory[m.execRunBlockID], record)
	m.execViewIndex[m.execRunBlockID] = len(m.execHistory[m.execRunBlockID]) - 1
	m.execViewBlockID = m.execRunBlockID
	m.execFollowTail = true
	m.execLogScroll = 0
	m.execCancel = func() {
		m.closeShellSession(session.key)
	}

	if err := session.run(script, m.execMsgCh, m.execStartedAt); err != nil {
		m.execRunning = false
		return func() tea.Msg { return execDoneMsg{err: err, exitCode: -1} }
	}
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

func commandForLanguage(lang, script string, session bool) (string, []string) {
	switch strings.ToLower(lang) {
	case "bash", "sh", "zsh", "shell":
		if session {
			return "bash", []string{"--noprofile", "--norc"}
		}
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

func supportsSession(lang string) bool {
	switch strings.ToLower(lang) {
	case "bash", "sh", "zsh", "shell":
		return true
	default:
		return false
	}
}

func sessionKey(item OutlineItem) string {
	return strings.ToLower(item.Lang)
}

func (m *Model) getOrCreateShellSession(key, lang string) (*shellSession, error) {
	if session, ok := m.shellSessions[key]; ok {
		return session, nil
	}
	name, args := commandForLanguage(lang, "", true)
	cmd := exec.Command(name, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	session := &shellSession{
		key:   key,
		lang:  lang,
		cmd:   cmd,
		stdin: stdin,
	}
	m.shellSessions[key] = session
	go session.readStream(stdout, "stdout")
	go session.readStream(stderr, "stderr")
	go func() {
		_ = cmd.Wait()
		session.failActiveRun(execDoneMsg{err: errors.New("session ended"), exitCode: -1})
	}()
	return session, nil
}

func (m *Model) closeShellSession(key string) {
	session, ok := m.shellSessions[key]
	if !ok {
		return
	}
	delete(m.shellSessions, key)
	session.failActiveRun(execDoneMsg{killed: true, exitCode: -1, duration: time.Since(session.startedAt)})
	if session.cmd.Process != nil {
		_ = session.cmd.Process.Kill()
	}
}

func (m *Model) closeAllShellSessions() {
	for key := range m.shellSessions {
		m.closeShellSession(key)
	}
}

func (s *shellSession) run(script string, out chan tea.Msg, startedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentCh != nil {
		return errors.New("session busy")
	}
	s.currentCh = out
	s.startedAt = startedAt
	s.token = fmt.Sprintf("rundown-%d", time.Now().UnixNano())
	payload := script + "\n" +
		"__rundown_status=$?\n" +
		fmt.Sprintf("printf '__RUNDOWN_DONE__ %s %%s\\n' \"$__rundown_status\"\n", s.token)
	_, err := io.WriteString(s.stdin, payload)
	return err
}

func (s *shellSession) readStream(r io.Reader, stream string) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 2*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if stream == "stdout" && s.handleSentinel(line) {
			continue
		}
		s.mu.Lock()
		ch := s.currentCh
		s.mu.Unlock()
		if ch != nil {
			ch <- execLineMsg{line: line, stream: stream}
		}
	}
}

func (s *shellSession) handleSentinel(line string) bool {
	if !strings.HasPrefix(line, "__RUNDOWN_DONE__ ") {
		return false
	}
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentCh == nil || parts[1] != s.token {
		return true
	}
	exitCode, _ := strconv.Atoi(parts[2])
	s.currentCh <- execDoneMsg{exitCode: exitCode, duration: time.Since(s.startedAt)}
	s.currentCh = nil
	s.token = ""
	return true
}

func (s *shellSession) failActiveRun(msg execDoneMsg) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentCh == nil {
		return
	}
	if msg.duration == 0 {
		msg.duration = time.Since(s.startedAt)
	}
	s.currentCh <- msg
	s.currentCh = nil
	s.token = ""
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
