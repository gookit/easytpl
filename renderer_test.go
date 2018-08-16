package view

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"bytes"
)

func TestRenderer_String(t *testing.T) {
	art := assert.New(t)
	r := NewRenderer()
	r.MustInitialize()

	bf := new(bytes.Buffer)

	err := r.String(bf, `hello {{.}}`, "tom")
	art.Nil(err)
	art.Equal("hello tom", bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{. | ucFirst}}`, "tom")
	art.Nil(err)
	art.Equal("hello Tom", bf.String())
}