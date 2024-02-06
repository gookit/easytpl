package strtpl

import (
	"github.com/gookit/goutil/strutil/textutil"
)

// StrTemplate implement a simple string template.
// Alias of textutil.LiteTemplate
type StrTemplate struct {
	textutil.LiteTemplate
}

// NewStrTemplate instance
func NewStrTemplate(opFns ...func(st *StrTemplate)) *StrTemplate {
	st := &StrTemplate{}
	for _, fn := range opFns {
		fn(st)
	}
	return st
}
