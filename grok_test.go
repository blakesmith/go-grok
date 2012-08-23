package grok

import (
	"testing"
)

func TestNew(t *testing.T) {
	g := New()
	if g == nil && g.g == nil {
		t.Fatal("Failed to initialize grok")
	}
}

func TestDayCompile(t *testing.T) {
	g := New()
	pattern := "%{DAY}"
	err := g.Compile(pattern)
	if err != nil {
		t.Fatal("Error:", err)
	}
}

func TestDayCompileAndMatch(t *testing.T) {
	g := New()
	g.AddPatternsFromFile("./patterns/base")
	text := "Tue May 15 11:21:42 [conn1047685] moveChunk deleted: 7157"
	pattern := "%{DAY}"
	err := g.Compile(pattern)
	if err != nil {
		t.Fatal("Error:", err)
	}
	match := g.Match(text)
	if match == nil {
		t.Fatal("Unable to match!")
	}
}
