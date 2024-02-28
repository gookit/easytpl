package easytpl

import (
	"bytes"
	"fmt"
	"sync"
)

func panicf(format string, args ...any) {
	if len(args) > 0 {
		panic("easyTpl: [ERROR] " + fmt.Sprintf(format, args...))
	}

	// only error message
	panic("easyTpl: [ERROR] " + format)
}

func panicErr(err error) {
	if err != nil {
		panic("easyTpl: [ERROR] " + err.Error())
	}
}

// match '{{ extend "parent.tpl" }}'
// var extendsRegex = regexp.MustCompile(`{{ *?extends +?"(.+?)" *?}}`)
var extendsBytes = []byte("extends ")

// parse line '{{ extend "parent.tpl" }}' and get "parent.tpl"
func getExtendsTplName(line []byte, td TplDelims) (string, bool) {
	line = bytes.TrimRight(line, " \t")

	if bytes.HasPrefix(line, []byte(td.Left)) &&
		bytes.HasSuffix(line, []byte(td.Right)) &&
		bytes.Contains(line, extendsBytes) {
		leftLen, rightLen := len(td.Left), len(td.Right)

		// remove left and right delimiters and spaces
		content := bytes.Trim(line[leftLen:len(line)-rightLen], " \"'")
		// remove "extends " prefix
		return string(bytes.TrimLeft(content[8:], " \"'")), true
	}

	return "", false
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
		New: func() any {
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
