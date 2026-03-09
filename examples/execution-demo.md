# Rundown Execution Demo (Safe)

This file is for testing code-block execution in `rundown`.
All commands are local and harmless by default.

## How To Use

1. Open this file in rundown:

### Third level heading

This is what it looks like

```bash
rundown examples/execution-demo.md
```

2. Switch to outline with `tab`.
3. Move to an executable block (`n` / `p` or `j` / `k`).
4. Press `r` to run.
5. Watch the log panel at the bottom.
6. Press `s` (or `Ctrl+X`) to stop a running command.

## Bash: Simple Output

```bash
echo "hello from bash"
echo "current shell: $SHELL"
echo "date: $(date '+%Y-%m-%d %H:%M:%S')"
```

## Ruby: Simple Output

```ruby
puts "hello from ruby"
puts "ruby version: #{RUBY_VERSION}"
puts "time: #{Time.now}"
```

## Bash: Streaming Logs

```bash
for i in 1 2 3 4 5; do
  echo "tick $i"
  sleep 0.3
done
echo "done"
```

## Ruby: Streaming Logs

```ruby
5.times do |i|
  puts "ruby tick #{i + 1}"
  sleep 0.3
end
puts "ruby done"
```

## Bash: Long Run (Use Stop)

Use this to test cancellation via `s` / `Ctrl+X`.

```bash
echo "starting long task"
for i in $(seq 1 120); do
  echo "line $i"
  sleep 0.5
done
echo "finished"
```

## Ruby: Long Run (Use Stop)

```ruby
puts "starting long ruby task"
120.times do |i|
  puts "ruby line #{i + 1}"
  sleep 0.5
end
puts "finished ruby"
```
