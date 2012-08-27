package grok

/*
#cgo LDFLAGS: -lgrok
#include <grok.h>
*/
import "C"

import (
	"errors"
	"fmt"
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
}

type Match struct {
	gm      C.grok_match_t
	Grok    *Grok
	Subject string
}

func New() *Grok {
	grok := new(Grok)
	grok.g = C.grok_new()
	if grok.g == nil {
		return nil
	}

	return grok
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

func (grok *Grok) Compile(pattern string) error {
	p := C.CString(pattern)
	defer C.free(unsafe.Pointer(p))

	ret := C.grok_compile(grok.g, p)
	if ret != GROK_OK {
		return errors.New(fmt.Sprintf("Failed to compile: %s", C.GoString(grok.g.errstr)))
	}

	return nil
}

func (grok *Grok) Match(text string) *Match {
	t := C.CString(text)

	var cmatch C.grok_match_t

	ret := C.grok_exec(grok.g, t, &cmatch)
	if ret != GROK_OK {
		return nil
	}

	match := new(Match)
	match.gm = cmatch
	match.Subject = C.GoString(cmatch.subject)
	match.Grok = grok

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

func (match *Match) Captures() map[string][]string {
	captures := make(map[string][]string)

	var name, substring *C.char
	var namelen, sublen C.int

	C.grok_match_walk_init(&match.gm)

	for C.grok_match_walk_next(&match.gm, &name, &namelen, &substring, &sublen) == GROK_OK {
		var substrings []string

		gname := C.GoStringN(name, namelen)
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
