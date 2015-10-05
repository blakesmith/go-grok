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

	g.AddPatternsFromFile("../patterns/base")
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
	if &match.gm == nil {
		t.Fatal("Match object not correctly populated")
	}
}

func TestMatchCaptures(t *testing.T) {
	g := New()
	defer g.Free()

	g.AddPatternsFromFile("../patterns/base")
	text := "Tue May 15 11:21:42 [conn1047685] moveChunk deleted: 7157"
	pattern := "%{DAY}"
	g.Compile(pattern)
	match := g.Match(text)
	if match == nil {
		t.Fatal("Unable to find match!")
	}

	captures := match.Captures()
	if dayCap := captures["DAY"][0]; dayCap != "Tue" {
		t.Fatal("Day should equal Tue")
	}
}

func TestURICaptures(t *testing.T) {
	g := New()
	defer g.Free()

	g.AddPatternsFromFile("../patterns/base")
	text := "https://www.google.com/search?q=moose&sugexp=chrome,mod=16&sourceid=chrome&ie=UTF-8"
	pattern := "%{URI}"
	g.Compile(pattern)
	match := g.Match(text)
	if match == nil {
		t.Fatal("Unable to find match!")
	}

	captures := match.Captures()

	if host := captures["URIHOST"][0]; host != "www.google.com" {
		t.Fatal("URIHOST should be www.google.com")
	}
	if path := captures["URIPATH"][0]; path != "/search" {
		t.Fatal("URIPATH should be /search")
	}
}

func TestDiscovery(t *testing.T) {
	g := New()
	defer g.Free()

	g.AddPattern("IP", "(?<![0-9])(?:(?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})[.](?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})[.](?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})[.](?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2}))(?![0-9])")

	text := "1.2.3.4"
	discovery := g.Discover(text)
	g.Compile(discovery)
	captures := g.Match(text).Captures()
	if ip := captures["IP"][0]; ip != text {
		t.Fatal("IP should be 1.2.3.4")
	}
}

func TestPileMatching(t *testing.T) {
	p := NewPile()
	defer p.Free()

	p.AddPattern("foo", ".*(foo).*")
	p.AddPattern("bar", ".*(bar).*")

	p.Compile("%{bar}")

	grok, match := p.Match("bar")

	captures := match.Captures()
	if bar := captures["bar"][0]; bar != "bar" {
		t.Fatal("Should match the bar pattern")
	}

	captures = grok.Match("bar").Captures()
	if bar := captures["bar"][0]; bar != "bar" {
		t.Fatal("Should match the bar pattern")
	}
}

func TestPileAddPatternsFromFile(t *testing.T) {
	p := NewPile()
	defer p.Free()

	p.AddPatternsFromFile("../patterns/base")
	p.Compile("%{DAY}")

	text := "Tue May 15 11:21:42 [conn1047685] moveChunk deleted: 7157"

	_, match := p.Match(text)

	captures := match.Captures()
	if day := captures["DAY"][0]; day != "Tue" {
		t.Fatal("Should match the Tue")
	}
}

/* Get the index of the first match in the string */
func TestMatchIndices(t *testing.T) {
	text := "Tue May 15 11:21:42 [conn1047685] moveChunk deleted: May 7157"
	g := New()
	g.Compile("May")

	match := g.Match(text)
	
	idx := match.FindIndex()
	if idx[0] != 4 {
		t.Fatal("Expected starting index 4, got", idx[0])
	}
	if idx[1] != 7 {
		t.Fatal("Expected end  index 7, got", idx[1])
	}
}

/* Support PCRE named captures: they can't start with `_`, and they're
    prefixed with `:` */
func TestPCRENamedCaptures(t *testing.T) {
	g := New()
	defer g.Free()

	g.AddPatternsFromFile("../patterns/base")
	text := "message - Tue November 2000 ALLCAPSHOST 12345"
	pattern := "(?P<word>[a-z]*) - %{DAY} %{MONTH} (?P<year>[0-9]*) (?P<host>[A-Z]*) %{BASE10NUM}"
	g.Compile(pattern)
	match := g.Match(text)
	if match == nil {
		t.Fatal("Unable to find match!")
	}

	captures := match.Captures()

	if host := captures[":word"][0]; host != "message" {
		t.Fatal("word should be 'message'")
	}
	if len(captures["DAY"]) != 1 {
		t.Fatal("Expected one group named DAY")
	}
	if path := captures["DAY"][0]; path != "Tue" {
		t.Fatal("DAY should be 'Tue'")
	}
	if len(captures["MONTH"]) != 1 {
		t.Fatal("Expected one group named MONTH")
	}
	if month := captures["MONTH"][0]; month != "November" {
		t.Fatal("month should be 'November'")
	}
	if len(captures[":year"]) != 1 {
		t.Fatal("Expected one group named year")
	}
        if year := captures[":year"][0]; year != "2000" {
		t.Fatal("year should be '2000'")
	}
	if len(captures[":host"]) != 1 {
		t.Fatal("Expected one group named host")
	}
        if host := captures[":host"][0]; host != "ALLCAPSHOST" {
		t.Fatal("host should be 'ALLCAPSHOST'")
	}
	if len(captures["BASE10NUM"]) != 1 {
		t.Fatal("Expected one group named BASE10NUM")
	}
        if num := captures["BASE10NUM"][0]; num != "12345" {
		t.Fatal("BASE10NUM should be '12345'")
	}
}

