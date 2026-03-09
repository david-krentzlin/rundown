# Rundown Demo: Public API Workflow (Bash + Ruby)

This walkthrough is intentionally long so the left pane requires scrolling.
It also contains multiple executable blocks in both Bash and Ruby.

## Scenario

We want to inspect public repository activity from the GitHub API and build
small summaries we can use in a terminal report.

Goals:

- Fetch repository metadata
- Extract commit information
- Compute a lightweight summary in Ruby
- Produce a final markdown report snippet

## Prerequisites

You should have these tools available:

- `bash`
- `curl`
- `jq` (optional but useful)
- `ruby`

No authentication is required for this demo, but rate limits are lower for
anonymous requests.

## Step 1: Pick a Repository

For this demo we use the public repository `charmbracelet/bubbletea`.

```bash
OWNER="charmbracelet"
REPO="bubbletea"
BASE="https://api.github.com/repos/${OWNER}/${REPO}"
echo "Using ${OWNER}/${REPO}"
```

The API endpoint we will call first is:

- `GET /repos/{owner}/{repo}`

## Step 2: Fetch Repository Metadata (Bash)

```bash
curl -sS "$BASE" > /tmp/rundown_repo.json
jq '{name, full_name, stargazers_count, forks_count, open_issues_count, default_branch}' /tmp/rundown_repo.json
```

If `jq` is missing, inspect the raw JSON with:

```bash
head -n 40 /tmp/rundown_repo.json
```

## Step 3: Fetch Recent Commits (Bash)

```bash
curl -sS "$BASE/commits?per_page=20" > /tmp/rundown_commits.json
jq '.[0:5] | map({sha: .sha[0:7], author: .commit.author.name, date: .commit.author.date, message: .commit.message})' /tmp/rundown_commits.json
```

This gives us the latest commit samples for quick inspection.

## Step 4: Generate a Compact Summary (Ruby)

Now we parse the JSON files in Ruby and create a compact text summary.

```ruby
require "json"

repo = JSON.parse(File.read("/tmp/rundown_repo.json"))
commits = JSON.parse(File.read("/tmp/rundown_commits.json"))

messages = commits.first(20).map { |c| c.dig("commit", "message").to_s.lines.first.to_s.strip }
authors = commits.first(20).map { |c| c.dig("commit", "author", "name").to_s }

author_counts = authors.each_with_object(Hash.new(0)) { |name, h| h[name] += 1 }
top_authors = author_counts.sort_by { |name, count| [-count, name] }.first(5)

puts "Repository: #{repo["full_name"]}"
puts "Stars: #{repo["stargazers_count"]}"
puts "Forks: #{repo["forks_count"]}"
puts "Open issues: #{repo["open_issues_count"]}"
puts "Default branch: #{repo["default_branch"]}"
puts
puts "Top recent authors:"
top_authors.each { |name, count| puts "- #{name}: #{count}" }
puts
puts "Recent commit titles:"
messages.first(10).each_with_index { |msg, idx| puts "#{idx + 1}. #{msg}" }
```

## Step 5: Build a Markdown Report Fragment (Ruby)

```ruby
require "json"

repo = JSON.parse(File.read("/tmp/rundown_repo.json"))
commits = JSON.parse(File.read("/tmp/rundown_commits.json"))

titles = commits.first(8).map { |c| c.dig("commit", "message").to_s.lines.first.to_s.strip }

puts "## Repo Snapshot"
puts
puts "- Name: #{repo["full_name"]}"
puts "- Stars: #{repo["stargazers_count"]}"
puts "- Forks: #{repo["forks_count"]}"
puts "- Open issues: #{repo["open_issues_count"]}"
puts
puts "### Recent Commit Titles"
titles.each { |t| puts "- #{t}" }
```

## Step 6: Single-Command Pipeline (Bash + Ruby)

This combines fetch + summarize in one executable target.

```bash
OWNER="charmbracelet"
REPO="bubbletea"
BASE="https://api.github.com/repos/${OWNER}/${REPO}"

curl -sS "$BASE" > /tmp/rundown_repo.json
curl -sS "$BASE/commits?per_page=30" > /tmp/rundown_commits.json

ruby <<'RUBY'
require "json"
repo = JSON.parse(File.read("/tmp/rundown_repo.json"))
commits = JSON.parse(File.read("/tmp/rundown_commits.json"))

puts "# Quick API Summary"
puts "Repo: #{repo["full_name"]}"
puts "Stars: #{repo["stargazers_count"]}"
puts "Commits fetched: #{commits.size}"
puts
commits.first(5).each_with_index do |c, i|
  title = c.dig("commit", "message").to_s.lines.first.to_s.strip
  sha = c["sha"].to_s[0, 7]
  puts "#{i + 1}. #{sha} #{title}"
end
RUBY
```

## Notes on Rate Limits

GitHub API rate limits are strict for unauthenticated requests.
If you hit limits, wait and retry later, or use a token in your own setup.

## Navigation Stress Section

The following paragraphs are intentionally verbose and repetitive to guarantee
scrolling space for UI testing. You can use this section to verify wheel
scrolling and synchronization between markdown and outline panes.

Paragraph A: Terminal user interfaces benefit from deterministic keybindings.
When scrolling behavior is predictable, users can confidently navigate large
technical documents, especially when they include code blocks and operational
notes.

Paragraph B: Outline synchronization should map the currently visible markdown
region to a nearby heading so users always retain location context. This is
especially useful in long runbooks and incident procedures.

Paragraph C: Executable code blocks become more valuable when they are grouped
under meaningful headings. This allows quick filtering of actionable targets.

Paragraph D: Repeated movement through content should feel stable with both
keyboard and mouse input. Hybrid navigation improves usability in real-world
terminal workflows.

Paragraph E: Rendering markdown with style support helps users distinguish
headings, code fences, and plain text at a glance, reducing cognitive load.

Paragraph F: When demos are realistic and operationally grounded, they make it
simpler to validate tool behavior before integrating command execution features.

## End

You have reached the bottom of the demo file.

```bash
echo "rundown demo complete"
```
