package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func BenchmarkParseMarkdown(b *testing.B) {
	src := benchmarkMarkdown(b)
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_ = ParseMarkdown(src)
	}
}

func BenchmarkInitialRenderCold(b *testing.B) {
	src := benchmarkMarkdown(b)
	doc := ParseMarkdown(src)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		m := NewModel(doc, "bench.md")
		m.SetViewport(140, 40)
		_ = m.render()
	}
}

func BenchmarkInitialRenderWarm(b *testing.B) {
	src := benchmarkMarkdown(b)
	doc := ParseMarkdown(src)
	m := NewModel(doc, "bench.md")
	m.SetViewport(140, 40)
	_ = m.render() // prime caches

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = m.render()
	}
}

func benchmarkMarkdown(b *testing.B) string {
	b.Helper()
	path := filepath.Join("..", "..", "examples", "scroll-demo.md")
	data, err := os.ReadFile(path)
	if err != nil {
		b.Fatalf("read benchmark markdown %s: %v", path, err)
	}

	// Amplify input size so regressions are obvious in benchmarks.
	return strings.Repeat(string(data)+"\n", 8)
}
