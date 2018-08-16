package view

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRenderer_AddFunc(t *testing.T) {
	art := assert.New(t)

	r := NewRenderer()
	r.AddFunc("test1", func() {})
	art.Panics(func() {
		r.AddFunc("test2", "invalid")
	})
	r.MustInitialize()

	art.Panics(func() {
		r.AddFunc("test3", func() {})
	})
}

func TestRenderer_Initialize(t *testing.T) {
	art := assert.New(t)

	r := &Renderer{}
	r.AddFunc("test", func() {})
	r.AddFuncMap(map[string]interface{}{
		"test1": func() {},
	})

	err := r.Initialize()
	art.Nil(err)
	// re-init
	art.Nil(r.Initialize())

	r1 := NewRenderer()
	art.Panics(func() {
		bf := new(bytes.Buffer)
		r1.Render(bf, "", nil)
	})

	bf := new(bytes.Buffer)
	r = NewInitialized(func(r *Renderer) {
		r.Layout = "layout"
		r.ViewsDir = "testdata/admin"
	})

	err = r.Render(bf, "home", "tom")
	art.Nil(err)

	fmt.Println(bf.String())

	// including itself.
	art.Len(r.LoadedTemplates(), 5+1)

	r = NewInitialized(func(r *Renderer) {
		r.Layout = "layout"
		r.ViewsDir = "testdata"
	})

}

func TestRenderer_LoadByGlob(t *testing.T) {
	bf := new(bytes.Buffer)
	art := assert.New(t)

	r := NewInitialized(func(r *Renderer) {
		// r.Debug = true
	})
	r.LoadByGlob("testdata/*")
	// r.LoadByGlob("testdata/*.tpl")
	err := r.Render(bf, "hello", "tom")
	art.Error(err)
	bf.Reset()
	err = r.Render(bf, "testdata/hello", "tom")
	art.Nil(err)
	art.Equal("hello tom", bf.String())

	r = NewInitialized(func(r *Renderer) {
		// r.Debug = true
	})
	r.LoadByGlob("testdata/*", "testdata/")
	// r.LoadByGlob("testdata/*.tpl")
	bf.Reset()
	err = r.Render(bf, "hello", "tom")
	art.Nil(err)
	art.Equal("hello tom", bf.String())
}

func TestRenderer_String(t *testing.T) {
	art := assert.New(t)
	r := NewRenderer()
	r.MustInitialize()

	bf := new(bytes.Buffer)

	err := r.String(bf, `hello {{.}}`, "tom")
	art.Nil(err)
	art.Equal("hello tom", bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{.name}}`, M{"name": "tom"})
	art.Nil(err)
	art.Equal("hello tom", bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{.Name}}`, struct {
		Name string
	}{"tom"})
	art.Nil(err)
	art.Equal("hello tom", bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{. | ucFirst}}`, "tom")
	art.Nil(err)
	art.Equal("hello Tom", bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{.|raw}}`, "<i id=\"icon\"></i>")
	art.Nil(err)
	art.Equal(`hello <i id="icon"></i>`, bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{ yield }}`, "tom")
	art.Error(err)
	art.Contains(err.Error(), "yield called")

	bf.Reset()
	err = r.String(bf, `hello {{ current }}`, "tom")
	art.Nil(err)
	art.Equal(`hello `, bf.String())
}

func TestRenderer_LoadStrings(t *testing.T) {
	bf := new(bytes.Buffer)
	art := assert.New(t)
	r := NewRenderer(func(r *Renderer) {
		r.Debug = true
		r.Layout = "layout"
	})
	r.Initialize()

	r.LoadStrings(map[string]string{
		"layout":       `{{ include "header" . }}, at layout, {{ yield }}, {{ include "footer" }}`,
		"admin/layout": `{{ include "header" }}, at admin layout, {{ yield }}, {{ include "footer" }}`,
		"header":       `is header:{{.}}`,
		"footer":       `is footer:{{.}}`,
		"home":         `hello {{.name}}`,
		"admin/home":   `1 at {{current}}:{{.}}`,
		"admin/login":  `2 at {{current}}:{{.}}`,
	})

	r.LoadString("admin/reg", `main: hello {{.}}`)

	ldNames := r.LoadedNames()
	art.Contains(ldNames, `"home"`)
	art.Contains(ldNames, `"admin/reg"`)
	art.Contains(ldNames, `"admin/home"`)

	// use layout
	bf.Reset()
	err := r.Render(bf, "admin/home", "H")
	art.Nil(err)
	art.Equal(`is header:H, at layout, 1 at admin/home:H, is footer:`, bf.String())
	// fmt.Println(bf.String())
	// return

	bf.Reset()
	err = r.Render(bf, "admin/login", "L")
	art.Nil(err)
	art.Equal(`is header:L, at layout, 2 at admin/login:L, is footer:`, bf.String())

	// custom layout
	bf.Reset()
	err = r.Render(bf, "admin/login", "L", "admin/layout")
	art.Nil(err)
	art.Equal(`is header:, at admin layout, 2 at admin/login:L, is footer:`, bf.String())

	// disable layout by param
	bf.Reset()
	err = r.Render(bf, "admin/login", "L", "")
	art.Nil(err)
	art.Equal("2 at admin/login:L", bf.String())

	// not use layout
	bf.Reset()
	err = r.Partial(bf, "home", M{"name": "tom"})
	art.Nil(err)
	art.Equal("hello tom", bf.String())

}

func TestRenderer_Partial(t *testing.T) {
	bf := new(bytes.Buffer)
	art := assert.New(t)

	r := NewInitialized()

	err := r.Partial(bf, "not-exist", nil)
	art.Error(err)
}
