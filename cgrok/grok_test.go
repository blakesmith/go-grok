package grok

import (
	"sync"
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
	err := g.Compile(pattern, false)
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
	err := g.Compile(pattern, false)
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
	g.Compile(pattern, false)
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
	g.Compile(pattern, false)
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
	g.Compile(discovery, false)
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

	p.Compile("%{bar}", false)

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
	p.Compile("%{DAY}", false)

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
	g.Compile("May", false)

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
	g.Compile(pattern, false)
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

/* Test multiple goroutines using the same Grok concurrently - we use a separate iterator and PCRE vector per match now */
func TestConcurrentCaptures(t *testing.T) {
	g := New()
	defer g.Free()

	g.AddPatternsFromFile("../patterns/base")
	g.AddPattern("S3_REQUEST_LINE", "(?:%{WORD:verb} %{NOTSPACE:request}(?: HTTP/%{NUMBER:httpversion})?|%{DATA:rawrequest})")
	text1 := "1124412d476eb4e8c9b691cacfa51bb990eff8169c3337e0be688c1caf1bdaf0 releases.rocana.com [11/Apr/2015:03:27:40 +0000] 10.220.7.37 arn:aws:iam::368902385577:user/mark FC206D08A83F5300 REST.POST.UPLOADS scalingdata-0.7.0.tar.gz \"POST /releases.rocana.com/scalingdata-0.7.0.tar.gz?uploads HTTP/1.1\" 200 - 370 - 8 7 \"-\" \"S3Console/0.4\" -"
	text2 := "1124412d476eb4e8c9b691cacfa51bb990eff8169c3337e0be688c1caf1bdaf0 releases.rocana.com [24/Jul/2015:01:34:43 +0000] 135.23.112.88 - A2AD9CC02C12642F REST.HEAD.OBJECT 1.2.0/rocana-installer-1.2.0.bin.asc \"HEAD /1.2.0/rocana-installer-1.2.0.bin.asc HTTP/1.1\" 200 - - 836 7 - \"-\" \"curl/7.37.1\" -"
	pattern := "%{WORD:owner} %{NOTSPACE:bucket} \\[%{HTTPDATE:timestamp}\\] %{IP:clientip} %{NOTSPACE:requester} %{NOTSPACE:request_id} %{NOTSPACE:operation} %{NOTSPACE:key} (?:\"%{S3_REQUEST_LINE}\"|-) (?:%{INT:response}|-) (?:-|%{NOTSPACE:error_code}) (?:%{INT:bytes}|-) (?:%{INT:object_size}|-) (?:%{INT:request_time_ms}|-) (?:%{INT:turnaround_time_ms}|-) (?:%{QS:referrer}|-) (?:\"?%{QS:agent}\"?|-) (?:-|%{NOTSPACE:version_id})"
	g.Compile(pattern, false)
	var s sync.WaitGroup
	for i := 0 ; i< 10000; i++ {
		s.Add(1)	
		go func(){
			defer s.Done()
			for j := 0; j < 5; j++ {
				if i % 2 == 0 {
					match := g.Match(text1)
					if match == nil {
						t.Fatal("Unable to match string 1")
					}
					captures := match.Captures()
					if captures["HTTPDATE:timestamp"][0] != "11/Apr/2015:03:27:40 +0000" {
						t.Fatal("Got unexpected timestamp "+captures["HTTPDATE:timestamp"][0])
					}	
 					if captures["QS:agent"][0] != "\"S3Console/0.4\"" {
						t.Fatal("Got unexpected agent " + captures["QS:agent"][0])
					}
					if captures["INT:bytes"][0] != "370" {
						t.Fatal("Got unexpected bytes "+captures["INT:bytes"][0])
					}
					match.Free()
				} else {
					match := g.Match(text2)
					if match == nil {
						t.Fatal("Unable to match string 2")
					}
					captures := match.Captures()
					if captures["HTTPDATE:timestamp"][0] != "24/Jul/2015:01:34:43 +0000" {
						t.Fatal("Got unexpected timestamp "+captures["HTTPDATE:timestamp"][0])
					}	
 					if captures["QS:agent"][0] != "\"curl/7.37.1\"" {
						t.Fatal("Got unexpected agent "+captures["QS:agent"][0])
					}
					if captures["INT:bytes"][0] != "" {
						t.Fatal("Got unexpected bytes "+captures["INT:bytes"][0])
					}
					if captures["INT:object_size"][0] != "836" {
						t.Fatal("Got unexpected size "+captures["INT:object_size"][0])
					}
					match.Free()
				}
			}
		}()
	}
	s.Wait()
}

func TestRenamedOnly(t *testing.T) {
	g := New()
	defer g.Free()

	g.AddPatternsFromFile("../patterns/base")
	text := "message - Tue November 2000 ALLCAPSHOST 12345"
	pattern := "(?P<word>[a-z]*) - %{DAY:day} %{MONTH} (?P<year>[0-9]*) (?P<host>[A-Z]*) %{BASE10NUM:number}"
	g.Compile(pattern, true)
	match := g.Match(text)
	if match == nil {
		t.Fatal("Unable to find match!")
	}
	captures := make(map[string]string)
	match.CaptureIntoMap(captures)
	if len(captures) != 5 {
		t.Fatal("Expected 5 groups to be extracted")
	}
	if host := captures["word"]; host != "message" {
		t.Fatal("word should be 'message'")
	}
	if day := captures["day"]; day != "Tue" {
		t.Fatal("`day` should be 'Tue'")
	}
	if _, ok := captures["MONTH"]; ok {
		t.Fatal("`MONTH` should be ignored")
	}
	if year := captures["year"]; year != "2000" {
		t.Fatal("`year` should be '2000'")
	}
	if host := captures["host"]; host != "ALLCAPSHOST" {
		t.Fatal("`host` should be 'ALLCAPSHOST'")
	}
	if num := captures["number"]; num != "12345" {
		t.Fatal("`number` should be '12345'")
	}
}

func BenchmarkOldGrok(b *testing.B) {
	g := New()
	defer g.Free()

	g.AddPatternsFromFile("../patterns/base")
	g.AddPattern("S3_REQUEST_LINE", "(?:%{WORD:verb} %{NOTSPACE:request}(?: HTTP/%{NUMBER:httpversion})?|%{DATA:rawrequest}) (?P<pcre_named>.*)")
	text := "1124412d476eb4e8c9b691cacfa51bb990eff8169c3337e0be688c1caf1bdaf0 releases.rocana.com [11/Apr/2015:03:27:40 +0000] 10.220.7.37 arn:aws:iam::368902385577:user/mark FC206D08A83F5300 REST.POST.UPLOADS scalingdata-0.7.0.tar.gz \"POST /releases.rocana.com/scalingdata-0.7.0.tar.gz?uploads HTTP/1.1\" 200 - 370 - 8 7 \"-\" \"S3Console/0.4\" -"
	pattern := "%{WORD:owner} %{NOTSPACE:bucket} \\[%{HTTPDATE:timestamp}\\] %{IP:clientip} %{NOTSPACE:requester} %{NOTSPACE:request_id} %{NOTSPACE:operation} %{NOTSPACE:key} (?:\"%{S3_REQUEST_LINE}\"|-) (?:%{INT:response}|-) (?:-|%{NOTSPACE:error_code}) (?:%{INT:bytes}|-) (?:%{INT:object_size}|-) (?:%{INT:request_time_ms}|-) (?:%{INT:turnaround_time_ms}|-) (?:%{QS:referrer}|-) (?:\"?%{QS:agent}\"?|-) (?:-|%{NOTSPACE:version_id})"
	g.Compile(pattern, false)
	for i := 0; i < b.N; i++ {
		m := g.Match(text)
		m.Captures()
		m.Free()
	}
}

func BenchmarkNewGrok(b *testing.B) {
	g := New()
	defer g.Free()

	g.AddPatternsFromFile("../patterns/base")
	g.AddPattern("S3_REQUEST_LINE", "(?:%{WORD:verb} %{NOTSPACE:request}(?: HTTP/%{NUMBER:httpversion})?|%{DATA:rawrequest}) (?P<pcre_named>.*)")
	text := "1124412d476eb4e8c9b691cacfa51bb990eff8169c3337e0be688c1caf1bdaf0 releases.rocana.com [11/Apr/2015:03:27:40 +0000] 10.220.7.37 arn:aws:iam::368902385577:user/mark FC206D08A83F5300 REST.POST.UPLOADS scalingdata-0.7.0.tar.gz \"POST /releases.rocana.com/scalingdata-0.7.0.tar.gz?uploads HTTP/1.1\" 200 - 370 - 8 7 \"-\" \"S3Console/0.4\" -"
	pattern := "%{WORD:owner} %{NOTSPACE:bucket} \\[%{HTTPDATE:timestamp}\\] %{IP:clientip} %{NOTSPACE:requester} %{NOTSPACE:request_id} %{NOTSPACE:operation} %{NOTSPACE:key} (?:\"%{S3_REQUEST_LINE}\"|-) (?:%{INT:response}|-) (?:-|%{NOTSPACE:error_code}) (?:%{INT:bytes}|-) (?:%{INT:object_size}|-) (?:%{INT:request_time_ms}|-) (?:%{INT:turnaround_time_ms}|-) (?:%{QS:referrer}|-) (?:\"?%{QS:agent}\"?|-) (?:-|%{NOTSPACE:version_id})"
	g.Compile(pattern, true)
	for i := 0; i < b.N; i++ {
		m := g.Match(text)
		m.Captures()
		m.Free()
	}
}

func BenchmarkNewGrokIntoMap(b *testing.B) {
	g := New()
	defer g.Free()

	g.AddPatternsFromFile("../patterns/base")
	g.AddPattern("S3_REQUEST_LINE", "(?:%{WORD:verb} %{NOTSPACE:request}(?: HTTP/%{NUMBER:httpversion})?|%{DATA:rawrequest}) (?P<pcre_named>.*)")
	text := "1124412d476eb4e8c9b691cacfa51bb990eff8169c3337e0be688c1caf1bdaf0 releases.rocana.com [11/Apr/2015:03:27:40 +0000] 10.220.7.37 arn:aws:iam::368902385577:user/mark FC206D08A83F5300 REST.POST.UPLOADS scalingdata-0.7.0.tar.gz \"POST /releases.rocana.com/scalingdata-0.7.0.tar.gz?uploads HTTP/1.1\" 200 - 370 - 8 7 \"-\" \"S3Console/0.4\" -"
	pattern := "%{WORD:owner} %{NOTSPACE:bucket} \\[%{HTTPDATE:timestamp}\\] %{IP:clientip} %{NOTSPACE:requester} %{NOTSPACE:request_id} %{NOTSPACE:operation} %{NOTSPACE:key} (?:\"%{S3_REQUEST_LINE}\"|-) (?:%{INT:response}|-) (?:-|%{NOTSPACE:error_code}) (?:%{INT:bytes}|-) (?:%{INT:object_size}|-) (?:%{INT:request_time_ms}|-) (?:%{INT:turnaround_time_ms}|-) (?:%{QS:referrer}|-) (?:\"?%{QS:agent}\"?|-) (?:-|%{NOTSPACE:version_id})"
	g.Compile(pattern, true)
	attr := make(map[string]string)
	for i := 0; i < b.N; i++ {
		m := g.Match(text)
		m.CaptureIntoMap(attr)
		m.Free()
	}
}
