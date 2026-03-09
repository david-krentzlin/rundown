package tui

import "testing"

func TestParseMarkdownBuildsOutlineWithHeadingsAndExecBlocks(t *testing.T) {
	src := "# Title\n\nText\n\n## Setup\n```bash\necho hi\n```\n\n### Run\n```python\nprint(1)\n```\n"
	doc := ParseMarkdown(src)

	if len(doc.Lines) == 0 {
		t.Fatal("expected parsed lines")
	}

	if got := len(doc.Outline); got != 5 {
		t.Fatalf("outline length = %d, want 5", got)
	}

	if doc.Outline[0].Kind != NodeHeading || doc.Outline[0].Title != "Title" {
		t.Fatalf("first outline item = %#v", doc.Outline[0])
	}

	if doc.Outline[2].Kind != NodeExec || doc.Outline[2].Lang != "bash" {
		t.Fatalf("expected bash exec node at index 2, got %#v", doc.Outline[2])
	}

	if doc.Outline[4].Kind != NodeExec || doc.Outline[4].Lang != "python" {
		t.Fatalf("expected python exec node at index 4, got %#v", doc.Outline[4])
	}
	for i, item := range doc.Outline {
		if item.ID == "" {
			t.Fatalf("outline[%d] id is empty", i)
		}
	}
}

func TestParseMarkdownExecBlockParsesSessionOption(t *testing.T) {
	src := "# Title\n```bash session=true\necho hi\n```\n"
	doc := ParseMarkdown(src)

	if len(doc.Outline) < 2 {
		t.Fatalf("outline length = %d, want at least 2", len(doc.Outline))
	}
	exec := doc.Outline[1]
	if exec.Kind != NodeExec {
		t.Fatalf("kind = %v, want NodeExec", exec.Kind)
	}
	if exec.Lang != "bash" {
		t.Fatalf("lang = %q, want bash", exec.Lang)
	}
	if !exec.Session {
		t.Fatal("expected session=true to enable persistent session mode")
	}
}
