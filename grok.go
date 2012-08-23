package grok

/*
#cgo LDFLAGS: -lgrok
#include <grok.h>
#include <grok_pattern.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

const (
	GROK_OK = iota
)

type Grok struct {
	g *C.grok_t
}

type Match struct {
	gm *C.grok_match_t
}

func New() *Grok {
	grok := new(Grok)
	grok.g = C.grok_new()
	if grok.g == nil {
		return nil
	}

	return grok
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

	ret := C.grok_compilen(grok.g, p, C.int(len(pattern)))
	if ret != GROK_OK {
		return errors.New(fmt.Sprintf("Failed to compile: %s", C.GoString(grok.g.errstr)))
	}

	return nil
}

func (grok *Grok) Match(text string) *Match {
	t := C.CString(text)
	defer C.free(unsafe.Pointer(t))
	var cmatch *C.grok_match_t

	ret := C.grok_execn(grok.g, t, C.int(len(text)), cmatch)
	if ret != GROK_OK {
		return nil
	}

	match := new(Match)
	match.gm = cmatch

	return match
}

func (grok *Grok) Free() {
	C.grok_free(grok.g)
}
