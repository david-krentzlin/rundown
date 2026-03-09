package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/david-krentzlin/rundown/internal/app"
	"github.com/david-krentzlin/rundown/internal/tui"
)

func main() {
	fileName := "untitled.md"
	content := []byte(defaultMarkdown())
	if len(os.Args) > 1 {
		fileName = os.Args[1]
		data, err := os.ReadFile(fileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", fileName, err)
			os.Exit(1)
		}
		content = data
	}

	model := tui.NewModel(tui.ParseMarkdown(string(content)), fileName)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s failed: %v\n", app.DisplayName(), err)
		os.Exit(1)
	}
}

func defaultMarkdown() string {
	return `# Rundown

Welcome to rundown.

Open a markdown file via:

    rundown path/to/file.md

Try the bundled scroll demo:

    rundown examples/scroll-demo.md

## Navigation
- Tab switches panes.
- Use hjkl in markdown.
- Use j/k and c/C e/E x n p in outline.

## Example
` + "```bash" + `
echo "hello from rundown"
` + "```" + `
`
}
