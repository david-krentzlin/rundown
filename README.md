# rundown

`rundown` is a terminal UI for reading Markdown that allows you to navigate using outline and execute code blocks in the file.

<img width="3426" height="1369" alt="image" src="https://github.com/user-attachments/assets/c35ac98a-f666-47e0-bb3c-2771d94b4275" />

<img width="3417" height="1362" alt="image" src="https://github.com/user-attachments/assets/bb6f3852-0a6f-4f53-99ad-5817a60189ea" />

## Build with AI

This tool has been built using agentic coding.  I wanted to validate the approach first.
The quality of the code is subpar as a result of that. However as far as I can tell this does not
inhibit functionality. I will polish this in the future.

## Requirements

- Go 1.26
- `mise` (optional, recommended)

Tool version pinning is defined in [`mise.toml`](./mise.toml).

## Install

Install the latest release with Go:

```bash
go install github.com/david-krentzlin/rundown/cmd/rundown@latest
```

This places the `rundown` binary in your `GOBIN` (or `$(go env GOPATH)/bin`).
Make sure that directory is on your `PATH`.

## Build

```bash
make build
```

Binary output:

```text
./bin/rundown
```

## Run

Run with built-in default content:

```bash
make run
```

Run with a file:

```bash
./bin/rundown path/to/file.md
```

Run the bundled scroll demo:

```bash
make run-demo
```

Demo file:

```text
examples/scroll-demo.md
```

Run the safe execution demo:

```bash
./bin/rundown examples/execution-demo.md
```

## Keyboard shortcuts

In-app help:

- `?` open/close help overlay
- `Esc` close help overlay

Global:

- `tab` switch focus between markdown and outline panes
- `Ctrl+A` jump to start of document
- `Ctrl+E` jump to end of document
- `Ctrl+C`, `Ctrl+Q`, `Q` quit (running commands are stopped first)

Markdown pane:

- `j` / `k` move down/up
- `h` / `l` fallback left/right navigation
- `J` / `K` jump next/previous heading
- `H` / `L` jump parent/first child heading
- mouse wheel scroll (when pointer is over the left pane)

Outline pane:

- `j` / `k` move down/up
- `c` / `C` collapse current/all headings
- `e` / `E` expand current/all headings
- `x` toggle executable-only outline
- `n` / `p` jump next/previous executable target
- `r` run selected executable target
- `s` stop running command

Execution/log panel:

- `v` show/hide panel
- `[` / `]` page previous/next run for selected executable
- `PgUp` / `PgDn` scroll logs
- `Ctrl+U` / `Ctrl+D` scroll logs
- `Home` / `End` jump to top/bottom of logs
- mouse wheel scroll (when pointer is over log panel)

Log auto-follow:

- While a command is running, logs follow output automatically.
- Manual upward scrolling pauses follow mode.
- `End` (or scrolling back to bottom) re-enables follow mode.

## Development

Run tests:

```bash
make test
```
