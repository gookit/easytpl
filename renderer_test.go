package easytpl_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gookit/easytpl"
	"github.com/gookit/goutil/testutil/assert"
)

func Example() {
	// equals to call: easytpl.NewRenderer() + r.MustInit()
	r := easytpl.NewInited(func(r *easytpl.Renderer) {
		// setting default layout
		r.Layout = "layout" // equals to "layout.tpl"
		// root dir. will autoload on init.
		r.ViewsDir = "testdata/layouts"
	})

	// fmt.Println(r.TemplateNames(true))
	bf := new(bytes.Buffer)

	// render template string
	_ = r.String(bf, `hello {{.}}`, "tom")
	fmt.Print(bf.String()) // hello tom

	// render template without layout
	_ = r.Partial(bf, "home", "tom")
	bf.Reset()

	// render with default layout
	_ = r.Render(bf, "home", "tom")
	bf.Reset()

	// render with custom layout
	_ = r.Render(bf, "home", "tom", "site/layout")
	bf.Reset()

	// load named template string
	r.LoadString("my-page", "welcome {{.}}")
	// now, you can use "my-page" as a template name
	_ = r.Partial(bf, "my-page", "tom") // welcome tom
	bf.Reset()

	// more ways for load root
	r.LoadByGlob("some/path/*", "some/path")
	r.LoadFiles("path/file1.tpl", "path/file2.tpl")
}

func TestRenderer_AddFunc(t *testing.T) {
	is := assert.New(t)

	r := easytpl.NewRenderer()
	r.AddFunc("test1", func() {})
	is.Panics(func() {
		r.AddFunc("test2", "invalid")
	})
	r.MustInit()

	is.Panics(func() {
		r.AddFunc("test3", func() {})
	})
}

func TestRenderer_LoadByGlob(t *testing.T) {
	bf := new(bytes.Buffer)
	is := assert.New(t)

	r := easytpl.NewInitialized(func(r *easytpl.Renderer) {
		// r.Debug = true
	})
	r.LoadByGlob("testdata/layouts/*")
	// r.LoadByGlob("testdata/layouts/*.tpl")
	err := r.Render(bf, "not-exist", "tom")
	is.Error(err)
	bf.Reset()

	err = r.Render(bf, "testdata/hello", "tom")
	is.Nil(err)
	is.Equal("hello tom", bf.String())

	r = easytpl.NewInitialized(func(r *easytpl.Renderer) {
		// r.Debug = true
	})
	r.LoadByGlob("testdata/*", "testdata/")
	// r.LoadByGlob("testdata/*.tpl")
	bf.Reset()

	err = r.Render(bf, "hello", "tom")
	is.Nil(err)
	is.Equal("hello tom", bf.String())
}

func TestRenderer_LoadFiles(t *testing.T) {
	is := assert.New(t)
	bf := new(bytes.Buffer)
	r := easytpl.NewInitialized()
	r.LoadFiles("testdata/hello.tpl")

	err := r.Render(bf, "testdata/hello", "tom")
	is.Nil(err)
	is.Equal("hello tom", bf.String())

	is.Panics(func() {
		r.LoadFiles("not-exist.tpl")
	})
}

func TestRenderer_String(t *testing.T) {
	is := assert.New(t)
	r := easytpl.NewRenderer()
	r.MustInit()

	bf := new(bytes.Buffer)

	err := r.String(bf, `hello {{.}}`, "tom")
	is.Nil(err)
	is.Equal("hello tom", bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{.name}}`, easytpl.M{"name": "tom"})
	is.Nil(err)
	is.Equal("hello tom", bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{.Name}}`, struct {
		Name string
	}{"tom"})
	is.Nil(err)
	is.Equal("hello tom", bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{. | upper}}`, "tom")
	is.Nil(err)
	is.Equal("hello TOM", bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{. | lower}}`, "TOM")
	is.Nil(err)
	is.Equal("hello tom", bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{. | ucFirst}}`, "tom")
	is.Nil(err)
	is.Equal("hello Tom", bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{. | ucFirst}}`, "")
	is.Nil(err)
	is.Equal("hello ", bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{.|raw}}`, "<i id=\"icon\"></i>")
	is.Nil(err)
	is.Equal(`hello <i id="icon"></i>`, bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{ yield }}`, "tom")
	is.Error(err)
	is.Contains(err.Error(), "yield called")

	bf.Reset()
	err = r.String(bf, `hello {{ current_tpl }}`, "tom")
	is.Nil(err)
	is.Equal(`hello `, bf.String())
}

func TestRenderer_LoadStrings(t *testing.T) {
	bf := new(bytes.Buffer)
	is := assert.New(t)
	r := easytpl.NewRenderer(easytpl.WithLayout("layout"))
	is.NoError(r.Initialize())

	r.LoadStrings(map[string]string{
		"layout":       `{{ include "header" . }}, at layout, {{ yield }}, {{ include "footer" }}`,
		"admin/layout": `{{ include "header" }}, at admin layout, {{ yield }}, {{ include "footer" }}`,
		"header":       `is header:{{.}}`,
		"footer":       `is footer:{{.}}`,
		"home":         `hello {{.name}}`,
		"admin/home":   `1 at {{current_tpl}}:{{.}}`,
		"admin/login":  `2 at {{current_tpl}}:{{ . }}`,
		"other":        `at {{current_tpl}}:{{ include "not-exist" }}`,
	})

	r.LoadString("admin/reg", `main: hello {{.}}`)

	ldNames := r.TemplateNames()
	is.Contains(ldNames, `"home"`)
	is.Contains(ldNames, `"admin/reg"`)
	is.Contains(ldNames, `"admin/home"`)

	// use layout
	bf.Reset()
	err := r.Render(bf, "admin/home", "H")
	is.Nil(err)
	is.Equal(`is header:H, at layout, 1 at admin/home:H, is footer:`, bf.String())
	// fmt.Println(bf.String())
	// return

	bf.Reset()
	err = r.Render(bf, "admin/login", "L")
	is.Nil(err)
	is.Equal(`is header:L, at layout, 2 at admin/login:L, is footer:`, bf.String())

	// custom layout
	bf.Reset()
	err = r.Render(bf, "admin/login", "L", "admin/layout")
	is.Nil(err)
	is.Equal(`is header:, at admin layout, 2 at admin/login:L, is footer:`, bf.String())

	// disable layout by param
	bf.Reset()
	err = r.Render(bf, "admin/login", "L", "")
	is.Nil(err)
	is.Equal("2 at admin/login:L", bf.String())

	// not use layout
	bf.Reset()
	err = r.Partial(bf, "home", easytpl.M{"name": "tom"})
	is.Nil(err)
	is.Equal("hello tom", bf.String())

	// include not exist
	bf.Reset()
	err = r.Partial(bf, "other", easytpl.M{"name": "tom"})
	is.Nil(err)
	is.Equal("at other:", bf.String())
}

func TestRenderer_Partial(t *testing.T) {
	bf := new(bytes.Buffer)
	is := assert.New(t)

	r := easytpl.NewInitialized()

	err := r.Partial(bf, "not-exist", nil)
	is.Error(err)
}
