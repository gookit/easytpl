package litetpl

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"github.com/gookit/goutil/basefn"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/strutil"
	"github.com/gookit/goutil/strutil/textutil"
)

// StrTemplate implement a simple string template
//
// - support replace vars
// - support pipeline filter handle
// - support default value
// - support custom func
type StrTemplate struct {
	textutil.VarReplacer
	// Funcs template funcs. refer the text/template.Funcs
	Funcs map[string]func(string) string
}

// NewStrTemplate instance
func NewStrTemplate(opFns ...func(st *StrTemplate)) *StrTemplate {
	st := &StrTemplate{}
	// st.WithFormat(defaultVarFormat)
	st.RenderFn = st.renderVars

	for _, fn := range opFns {
		fn(st)
	}
	return st
}

// Init StrTemplate
func (t *StrTemplate) Init() {
	if t.init {
		return
	}

	basefn.PanicIf(t.Right == "", "var format Right chars is required")

	t.lLen, t.rLen = len(t.Left), len(t.Right)
	t.varReg = regexp.MustCompile(regexp.QuoteMeta(t.Left) + `(.+)` + regexp.QuoteMeta(t.Right))
}

func (t *StrTemplate) renderVars(s string, varMap map[string]string) string {
	return t.varReg.ReplaceAllStringFunc(s, func(sub string) string {
		name := strings.TrimSpace(sub[t.lLen : len(sub)-t.rLen])

		var defVal string
		if t.parseDef && strings.ContainsRune(name, '|') {
			name, defVal = strutil.TrimCut(name, "|")
		}

		if val, ok := varMap[name]; ok {
			return val
		}

		if t.NotFound != nil {
			if val, ok := t.NotFound(name); ok {
				return val
			}
		}

		if len(defVal) > 0 {
			return defVal
		}
		t.missVars = append(t.missVars, name)
		return sub
	})
}

// ErrFuncNotFound error
var ErrFuncNotFound = errors.New("template func not found")

func (t *StrTemplate) applyFilters(val string, filters []string) (string, error) {
	// filters like: "trim|upper|substr:1,2" => ["trim", "upper", "substr:1,2"]
	for _, filter := range filters {
		if fn, ok := t.Funcs[filter]; ok {
			val = fn(val)
		} else {
			return "", ErrFuncNotFound
		}
	}

	return val, nil
}

var builtInFuncs = template.FuncMap{
	// don't escape content
	"raw": func(s string) string {
		return s
	},
	"trim": func(s string) string {
		return strings.TrimSpace(s)
	},
	// join strings
	"join": func(ss []string, sep string) string {
		return strings.Join(ss, sep)
	},
	// lower first char
	"lcFirst": func(s string) string {
		return strutil.LowerFirst(s)
	},
	// upper first char
	"upFirst": func(s string) string {
		return strutil.UpperFirst(s)
	},
}

var (
	errorType        = reflect.TypeOf((*error)(nil)).Elem()
	fmtStringerType  = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
	reflectValueType = reflect.TypeOf((*reflect.Value)(nil)).Elem()
)

// goodFunc reports whether the function or method has the right result signature.
func goodFunc(typ reflect.Type) bool {
	// We allow functions with 1 result or 2 results where the second is an error.
	switch {
	case typ.NumOut() == 1:
		return true
	case typ.NumOut() == 2 && typ.Out(1) == errorType:
		return true
	}
	return false
}

// TextRenderOpt render text template options
type TextRenderOpt struct {
	// Output use custom output writer
	Output io.Writer
	// Funcs add custom template functions
	Funcs template.FuncMap
}

// RenderOptFn render option func
type RenderOptFn func(opt *TextRenderOpt)

// NewRenderOpt create a new render options
func NewRenderOpt(optFns []RenderOptFn) *TextRenderOpt {
	opt := &TextRenderOpt{}
	for _, fn := range optFns {
		fn(opt)
	}
	return opt
}

// RenderTpl render go template string or file.
func RenderTpl(input string, data any, optFns ...RenderOptFn) string {
	return RenderGoTpl(input, data, optFns...)
}

// RenderGoTpl render input text or template file.
func RenderGoTpl(input string, data any, optFns ...RenderOptFn) string {
	opt := NewRenderOpt(optFns)

	t := template.New("text-renderer")
	t.Funcs(builtInFuncs)
	if len(opt.Funcs) > 0 {
		t.Funcs(opt.Funcs)
	}

	if !strings.Contains(input, "{{") && fsutil.IsFile(input) {
		template.Must(t.ParseFiles(input))
	} else {
		template.Must(t.Parse(input))
	}

	// use custom output writer
	if opt.Output != nil {
		basefn.MustOK(t.Execute(opt.Output, data))
		return "" // return empty string
	}

	// use buffer receive rendered content
	buf := new(bytes.Buffer)
	basefn.MustOK(t.Execute(buf, data))
	return buf.String()
}
