package grok

/*
#cgo CFLAGS: -I. -std=gnu99
#cgo windows LDFLAGS: -L. -lws2_32
#include "grok.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"
)

const (
	GROK_OK = iota
	GROK_ERROR_FILE_NOT_ACCESSIBLE
	GROK_ERROR_PATTERN_NOT_FOUND
	GROK_ERROR_UNEXPECTED_READ_SIZE
	GROK_ERROR_COMPILE_FAILED
	GROK_ERROR_UNINITIALIZED
	GROK_ERROR_PCRE_ERROR
	GROK_ERROR_NOMATCH
)

type Grok struct {
	g *C.grok_t
	stringCacheLock sync.RWMutex
	stringCache map[uintptr]string
}

type Match struct {
	gm C.grok_match_t
	grok *Grok
	subject string

	/* Iterator state - each match only supports one iterator at a time.
	   This saves up doing a few heap allocations per group within each match */
	gname string
	gsubstring string
	name *C.char
	namelen, suboffset, sublen C.int
}

type Pile struct {
	Patterns     map[string]string
	PatternFiles []string
	Groks        []*Grok
}

func New() *Grok {
	grok := new(Grok)
	grok.g = C.grok_new()
	if grok.g == nil {
		return nil
	}
	grok.stringCache = make(map[uintptr]string)
	return grok
}

/* Only look up group names, which live as long as the grok and are frequently reused. 
   With other strings there's a risk of getting the wrong result if memory has been freed and reused. */
func (grok *Grok) gostringn(str *C.char, len C.int) string {
	ptr := uintptr(unsafe.Pointer(str))
	grok.stringCacheLock.RLock()
	goStr, ok := grok.stringCache[ptr]
	grok.stringCacheLock.RUnlock()
	if ok {
		return goStr
	}
	goStr = C.GoStringN(str, len)
	grok.stringCacheLock.Lock()
	grok.stringCache[ptr] = goStr
	grok.stringCacheLock.Unlock()
	return goStr
}

func (grok *Grok) AddPattern(name, pattern string) {
	cname := C.CString(name)
	cpattern := C.CString(pattern)
	defer C.free(unsafe.Pointer(cname))
	defer C.free(unsafe.Pointer(cpattern))

	C.grok_pattern_add(grok.g, cname, C.strlen(cname), cpattern, C.strlen(cpattern))
}

func (grok *Grok) AddPatternsFromFile(path string) error {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	ret := C.grok_patterns_import_from_file(grok.g, cpath)
	if ret != GROK_OK {
		return errors.New(fmt.Sprintf("Failed to add path %s", path))
	}

	return nil
}

func (grok *Grok) Compile(pattern string, onlyRenamed bool) error {
	p := C.CString(pattern)
	defer C.free(unsafe.Pointer(p))

	ret := C.grok_compile(grok.g, p, C.int(boolToInt(onlyRenamed)))
	if ret != GROK_OK {
		return errors.New(fmt.Sprintf("Failed to compile: %s", C.GoString(grok.g.errstr)))
	}

	return nil
}

/* Note that Matches must be freed after use, to free the C string
   used for matching and the PCRE vector */
func (grok *Grok) Match(text string) *Match {
	t := C.CString(text)

	var cmatch C.grok_match_t

	ret := C.grok_exec(grok.g, t, &cmatch)
	if ret != GROK_OK {
		C.free(unsafe.Pointer(t))
		return nil
	}

	match := new(Match)
	match.gm = cmatch
	match.grok = grok
	match.subject = text	
	return match
}

func (grok *Grok) Discover(text string) string {
	ctext := C.CString(text)
	defer C.free(unsafe.Pointer(ctext))

	gdt := C.grok_discover_new(grok.g)
	var discovery *C.char
	var discoverylen C.int

	C.grok_discover(gdt, ctext, &discovery, &discoverylen)

	return C.GoStringN(discovery, discoverylen)
}

func (grok *Grok) Free() {
	C.grok_free(grok.g)
}

func NewPile() *Pile {
	pile := new(Pile)
	pile.Patterns = make(map[string]string)
	pile.PatternFiles = make([]string, 0)
	pile.Groks = make([]*Grok, 0)

	return pile
}

func (pile *Pile) Free() {
	for _, grok := range pile.Groks {
		grok.Free()
	}
}

func (pile *Pile) AddPattern(name, str string) {
	pile.Patterns[name] = str
}

func (pile *Pile) Compile(pattern string, onlyRenamed bool) error {
	grok := New()
	if grok == nil {
		return errors.New("Unable to initialize grok!")
	}

	for name, value := range pile.Patterns {
		grok.AddPattern(name, value)
	}

	for _, path := range pile.PatternFiles {
		if err := grok.AddPatternsFromFile(path); err != nil {
			return err
		}
	}

	grok.Compile(pattern, onlyRenamed)
	pile.Groks = append(pile.Groks, grok)

	return nil
}

func (pile *Pile) AddPatternsFromFile(path string) {
	pile.PatternFiles = append(pile.PatternFiles, path)
}

func (pile *Pile) Match(str string) (*Grok, *Match) {
	for _, grok := range pile.Groks {
		match := grok.Match(str)
		if match != nil {
			return grok, match
		}
	}

	return nil, nil
}

func (match *Match) Captures() map[string][]string {
	captures := make(map[string][]string)

	var name, substring *C.char
	var namelen, sublen C.int

	C.grok_match_walk_init(&match.gm)

	for C.grok_match_walk_next(&match.gm, &name, &namelen, &substring, &sublen) == GROK_OK {
		var substrings []string
		gname := match.grok.gostringn(name, namelen)
		gsubstring := C.GoStringN(substring, sublen)

		if val := captures[gname]; val == nil {
			substrings = make([]string, 0)
		} else {
			substrings = val
		}

		captures[gname] = append(substrings, gsubstring)
	}
	C.grok_match_walk_end(&match.gm)

	return captures
}

/* Start an iterator for the match results. The iterator must be free'd when no longer in use.
   Each match has state for one iterator, multiple concurrent iterators are not supported. */
func (match *Match) StartIterator() {
	C.grok_match_walk_init(&match.gm)
}

/* Try to advance the iterator. Returns false at the end of the list of groups. */
func (match *Match) Next() bool {
	if C.grok_match_walk_next_offsets(&match.gm, &match.name, &match.namelen, &match.suboffset, &match.sublen) != GROK_OK {
		return false
	}
	match.gname = match.grok.gostringn(match.name, match.namelen)
	match.gsubstring = ""
	/* Do some pointer arithmetic to find the index of the match within the string. The match should always
           be between the beginning of the subject and the end, or an empty string */
	if int(match.sublen) > 0 {
		if int(match.suboffset) > -1 && int(match.suboffset + match.sublen) <= len(match.subject) {
			match.gsubstring = match.subject[int(match.suboffset):int(match.suboffset + match.sublen)]
		}
	}
	return true
}

/* Get the current group name and substring match from the iterator */
func (match *Match) Group() (string, string) {
	return match.gname, match.gsubstring	
}

func (match *Match) EndIterator() {
	C.grok_match_walk_end(&match.gm)
}

func (match *Match) Free() {
	ptr := unsafe.Pointer(match.gm.subject)
	if uintptr(ptr) != 0 {
		C.free(ptr)
	}
	C.grok_match_free(&match.gm)
}

/* Returns an array of two integers, where the first is the starting index of the match, and 
   the second is the last index of the match. This is the same convention as the Golang regexp
   library's `FindIndex`. */
func (match *Match) FindIndex() []int {
	return []int{int(match.gm.start), int(match.gm.end)}
}

func boolToInt(b bool) int {
  if b {
    return 1
  }
  return 0
}
