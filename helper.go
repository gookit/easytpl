package view

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"sync"
)

// match '{{ extend "parent.tpl" }}'
// var extendsRegex = regexp.MustCompile(`{{ *?extends +?"(.+?)" *?}}`)
var globalFuncMap = template.FuncMap{
	// don't escape content
	"raw": func(s string) template.HTML {
		return template.HTML(s)
	},
	"trim":  strings.TrimSpace,
	"join":  strings.Join,
	"lower": strings.ToLower,
	"upper": strings.ToUpper,
	// uppercase first char
	"ucFirst": func(s string) string {
		if len(s) != 0 {
			f := s[0]
			// is lower
			if f >= 'a' && f <= 'z' {
				return strings.ToUpper(string(f)) + s[1:]
			}
		}

		return s
	},
	"yield": func() (string, error) {
		return "", fmt.Errorf("yield called with no layout defined")
	},
	// add a empty func for compile
	"current": func() string {
		return ""
	},
	"extends": func(name string) string {
		return ""
	},
}

// CleanExt will clean file ext.
// eg
// 		"some.tpl" -> "some"
// 		"path/some.tpl" -> "path/some"
func (r *Renderer) cleanExt(name string) string {
	if len(r.ExtNames) == 0 {
		return name
	}

	// has ext
	if pos := strings.LastIndexByte(name, '.'); pos > 0 {
		ext := name[pos:]
		if r.IsValidExt(ext) {
			return name[0:pos]
		}
	}

	return name
}

func (r *Renderer) getLayoutName(settings []string) string {
	var layout string

	disableLayout := r.DisableLayout
	if len(settings) > 0 {
		layout = strings.TrimSpace(settings[0])
		if layout == "" {
			disableLayout = true
		}
	} else {
		layout = r.Layout
	}

	// Apply layout
	if !disableLayout && layout != "" {
		return layout
	}
	return ""
}

func (r *Renderer) debugf(format string, args ...interface{}) {
	if r.Debug {
		fmt.Printf("view: [DEBUG] "+format+"\n", args...)
	}
}

func panicErr(err error) {
	if err != nil {
		panic("view: [ERROR] " + err.Error())
	}
}

/*************************************************************
 * buffer Pool
 *************************************************************/

// bufferPool A bufferPool is a type-safe wrapper around a sync.Pool.
type bufferPool struct {
	p *sync.Pool
}

// newBufferPool constructs a new bufferPool.
func newBufferPool() *bufferPool {
	return &bufferPool{&sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}}
}

// Get retrieves a Buffer from the pool, creating one if necessary.
func (bp bufferPool) get() *bytes.Buffer {
	buf := bp.p.Get().(*bytes.Buffer)
	return buf
}

func (bp bufferPool) put(buf *bytes.Buffer) {
	buf.Reset()
	bp.p.Put(buf)
}
