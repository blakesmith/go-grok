package grok

/*
#cgo LDFLAGS: -lgrok
#include <grok.h>
*/
import "C"

import (
	"errors"
	"fmt"
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

func (grok *Grok) Compile(pattern string) error {
	p := C.CString(pattern)
	ret := C.grok_compilen(grok.g, p, C.int(len(pattern)))
	if ret != GROK_OK {
		return errors.New(fmt.Sprintf("Failed to compile: %s", C.GoString(grok.g.errstr)))
	}

	return nil
}
