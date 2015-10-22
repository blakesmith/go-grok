package grok

/*
#cgo CFLAGS: -I. -std=gnu99
#cgo windows CFLAGS: -IC:/msys64/mingw64/include
#cgo windows LDFLAGS: -L. -lportablexdr -lws2_32
#include "grok.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"strings"
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
	stringCache map[uintptr]string
}

type Match struct {
	gm C.grok_match_t
	grok *Grok
	subject string
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
	goStr, ok := grok.stringCache[ptr]
	if ok {
		return goStr
	}
	goStr = C.GoStringN(str, len)
	grok.stringCache[ptr] = goStr
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

/* Add captures to the map, clobbering existing keys. Only extracts
   sub-expressions which have been explicitly renamed - `%{DAY}` will not be extracted,
   but `%{DAY:day}` will be extracted with the key `day`. Sub-expressions with duplicate names
   will clobber each other as well - the last match will remain. */
func (match *Match) CaptureIntoMap(captures map[string]string) {
	var name *C.char
	var namelen, suboffset, sublen C.int

	C.grok_match_walk_init(&match.gm)

	for C.grok_match_walk_next_offsets(&match.gm, &name, &namelen, &suboffset, &sublen) == GROK_OK {
		gname := match.grok.gostringn(name, namelen)
		colonIdx := strings.Index(gname, ":")
		if colonIdx > -1 {
			/* Do some pointer arithmetic to find the index of the match within the string. The match should always
                           be between the beginning of the subject and the end, or an empty string */
			var gsubstring string
			if int(sublen) > 0 {
				if int(suboffset) > -1 && int(suboffset + sublen) <= len(match.subject) {
					gsubstring = match.subject[int(suboffset):int(suboffset + sublen)]
				}
			}
			captures[gname[colonIdx+1:]] = gsubstring
		}
	}
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
