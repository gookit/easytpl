package view

import (
	"bytes"
	"sync"
)

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
