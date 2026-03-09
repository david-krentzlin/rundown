# rundown

`rundown` is a terminal UI for reading Markdown with a synchronized outline pane.
It is built with Bubble Tea v2 and uses Lip Gloss for layout/styling.

## Requirements

- Go 1.26
- `mise` (optional, recommended)

Tool version pinning is defined in [`mise.toml`](./mise.toml).

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

## Keybindings

Global:

- `tab` switch focus between markdown and outline panes
- `C-c`, `C-q`, `Q` quit

Markdown pane:

- `j` / `k` move down/up
- `h` / `l` move left/up-line and right/down-line fallback navigation
- `H` parent heading
- `J` next heading
- `K` previous heading
- `L` first child heading
- mouse wheel scroll (when cursor is over the left pane)

Outline pane:

- `j` / `k` move down/up
- `c` collapse current heading
- `C` collapse all headings
- `e` expand current heading
- `E` expand all headings
- `x` toggle executable-only view
- `n` next executable target
- `p` previous executable target
- `r` reserved for execution (no-op currently)

## Development

Run tests:

```bash
make test
```

Run deterministic repository checks:

```bash
make quality-gate
```
