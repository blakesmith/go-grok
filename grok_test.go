package grok

import (
	"testing"
)

func TestNew(t *testing.T) {
	g := New()
	defer g.Free()

	if g == nil && g.g == nil {
		t.Fatal("Failed to initialize grok")
	}
}

func TestDayCompile(t *testing.T) {
	g := New()
	defer g.Free()

	pattern := "%{DAY}"
	err := g.Compile(pattern)
	if err != nil {
		t.Fatal("Error:", err)
	}
}

func TestDayCompileAndMatch(t *testing.T) {
	g := New()
	defer g.Free()

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
	if match.gm == nil {
		t.Fatal("Match object not correctly populated")
	}

	if match.Subject != text {
		t.Fatal("Subject is equal to:", match.Subject)
	}
	if match.Grok != g {
		t.Fatal("Grok not correctly set")
	}
}
