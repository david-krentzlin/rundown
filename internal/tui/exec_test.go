package tui

import (
	"strings"
	"testing"
)

func TestSelectedExecutableBlockExtractsScript(t *testing.T) {
	doc := ParseMarkdown("# A\n```ruby\nputs 'hi'\nputs 'bye'\n```\n")
	m := NewModel(doc, "test.md")
	m.outlineIdx = 1

	item, script, ok := m.selectedExecutableBlock()
	if !ok {
		t.Fatal("expected executable block")
	}
	if item.Kind != NodeExec {
		t.Fatalf("kind = %v, want NodeExec", item.Kind)
	}
	want := "puts 'hi'\nputs 'bye'"
	if script != want {
		t.Fatalf("script = %q, want %q", script, want)
	}
}

func TestCommandForLanguage(t *testing.T) {
	tests := []struct {
		lang string
		name string
		arg0 string
	}{
		{lang: "bash", name: "bash", arg0: "-lc"},
		{lang: "ruby", name: "ruby", arg0: "-e"},
		{lang: "python", name: "python3", arg0: "-c"},
		{lang: "js", name: "node", arg0: "-e"},
	}
	for _, tt := range tests {
		name, args := commandForLanguage(tt.lang, "echo hi", false)
		if name != tt.name {
			t.Fatalf("%s name = %q, want %q", tt.lang, name, tt.name)
		}
		if len(args) == 0 || args[0] != tt.arg0 {
			t.Fatalf("%s args[0] = %q, want %q", tt.lang, firstOrEmpty(args), tt.arg0)
		}
	}
}

func TestCommandForLanguageSessionBash(t *testing.T) {
	name, args := commandForLanguage("bash", "echo hi", true)
	if name != "bash" {
		t.Fatalf("name = %q, want bash", name)
	}
	if len(args) != 2 || args[0] != "--noprofile" || args[1] != "--norc" {
		t.Fatalf("args = %v, want [--noprofile --norc] for session shell", args)
	}
}

func TestSessionTruePersistsShellStateAcrossBlocks(t *testing.T) {
	doc := ParseMarkdown("# One\n```bash session=true\nexport RUNDOWN_SESSION_VAR=kept\n```\n# Two\n```bash session=true\necho \"$RUNDOWN_SESSION_VAR\"\n```\n")
	m := NewModel(doc, "test.md")

	m.outlineIdx = 1
	cmd := m.runSelectedExecutable()
	for cmd != nil {
		msg := cmd()
		_, cmd = m.Update(msg)
		if !m.execRunning {
			break
		}
	}
	if m.execRunning {
		t.Fatal("expected first session run to finish")
	}

	m.outlineIdx = 3
	cmd = m.runSelectedExecutable()
	for cmd != nil {
		msg := cmd()
		_, cmd = m.Update(msg)
		if !m.execRunning {
			break
		}
	}
	defer m.closeAllShellSessions()

	blockID := m.doc.Outline[3].ID
	history := m.execHistory[blockID]
	if len(history) == 0 {
		t.Fatal("expected execution history for second session block")
	}
	found := false
	for _, line := range history[len(history)-1].Logs {
		if strings.TrimSpace(line.Text) == "kept" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected persisted shell variable in logs, got %#v", history[len(history)-1].Logs)
	}
}

func firstOrEmpty(v []string) string {
	if len(v) == 0 {
		return ""
	}
	return v[0]
}
