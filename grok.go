package grok

/*
#cgo LDFLAGS: -lgrok
#include <grok.h>
*/
import "C"

type Grok struct {
	g *C.grok_t
}

func New() *Grok {
	grok := new(Grok)
	grok.g = C.grok_new()
	if grok.g == nil {
		return nil
	}

	return grok
}