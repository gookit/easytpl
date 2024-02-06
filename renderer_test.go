package easytpl_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/gookit/easytpl"
	"github.com/gookit/goutil/testutil/assert"
)

func Example() {
	// equals to call: easytpl.NewRenderer() + r.MustInit()
	r := easytpl.NewInitialized(func(r *easytpl.Renderer) {
		// setting default layout
		r.Layout = "layout" // equals to "layout.tpl"
		// templates dir. will auto load on init.
		r.ViewsDir = "testdata"
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
	// now, you can use "my-page" as an template name
	_ = r.Partial(bf, "my-page", "tom") // welcome tom
	bf.Reset()

	// more ways for load templates
	r.LoadByGlob("some/path/*", "some/path")
	r.LoadFiles("path/file1.tpl", "path/file2.tpl")
}

func TestRenderer_AddFunc(t *testing.T) {
	art := assert.New(t)

	r := easytpl.NewRenderer()
	r.AddFunc("test1", func() {})
	art.Panics(func() {
		r.AddFunc("test2", "invalid")
	})
	r.MustInit()

	art.Panics(func() {
		r.AddFunc("test3", func() {})
	})
}

func TestRenderer_Initialize(t *testing.T) {
	art := assert.New(t)

	// r := &Renderer{}
	easytpl.AddFunc("test", func() string { return "" })
	art.Panics(func() {
		easytpl.LoadFiles("testdata/home.tpl")
	})
	art.Panics(func() {
		easytpl.LoadByGlob("testdata/site/*.tpl", "testdata/site")
	})
	easytpl.AddFuncMap(map[string]interface{}{
		"test1": func() string { return "" },
	})

	r := easytpl.Default()
	r.Debug = true

	art.NoError(r.Initialize())

	// re-init
	art.Nil(r.Initialize())
	art.NotNil(r.Templates())

	r.LoadByGlob("testdata/site/*.tpl", "testdata/site")
	art.Len(r.TemplateFiles(), 5)

	tpl := r.Template("header")
	art.NotNil(tpl)
	tpl = r.Template("header.tpl")
	art.NotNil(tpl)

	bf := new(bytes.Buffer)
	r1 := easytpl.NewRenderer()
	art.Panics(func() {
		_ = r1.Render(bf, "", nil)
	})

	// use layout
	r = easytpl.NewInitialized(func(r *easytpl.Renderer) {
		r.Layout = "layout"
		r.ViewsDir = "testdata/admin"
	})

	// including itself.
	art.Len(r.LoadedTemplates(), 5+1)

	bf.Reset()
	err := r.Render(bf, "home.tpl", "tom")
	art.Nil(err)
	str := bf.String()
	art.Contains(str, "admin header")
	art.Contains(str, "home: hello")
	art.Contains(str, "admin footer")

	art.Panics(func() {
		_ = r.Render(bf, "home.tpl", "tom", "not-exist.tpl")
	})

	r = easytpl.NewInitialized(func(r *easytpl.Renderer) {
		r.Layout = "layout"
		r.ViewsDir = "testdata"
	})

	ns := r.TemplateNames(true)
	art.Contains(ns, "header")
	art.Contains(ns, "admin/header")
	art.Contains(ns, "site/header")

	easytpl.Revert() // Revert
}

func TestRenderer_LoadByGlob(t *testing.T) {
	bf := new(bytes.Buffer)
	art := assert.New(t)

	r := easytpl.NewInitialized(func(r *easytpl.Renderer) {
		// r.Debug = true
	})
	r.LoadByGlob("testdata/*")
	// r.LoadByGlob("testdata/*.tpl")
	err := r.Render(bf, "not-exist", "tom")
	art.Error(err)
	bf.Reset()

	err = r.Render(bf, "testdata/hello", "tom")
	art.Nil(err)
	art.Equal("hello tom", bf.String())

	r = easytpl.NewInitialized(func(r *easytpl.Renderer) {
		// r.Debug = true
	})
	r.LoadByGlob("testdata/*", "testdata/")
	// r.LoadByGlob("testdata/*.tpl")
	bf.Reset()

	err = r.Render(bf, "hello", "tom")
	art.Nil(err)
	art.Equal("hello tom", bf.String())
}

func TestRenderer_LoadFiles(t *testing.T) {
	art := assert.New(t)
	bf := new(bytes.Buffer)
	r := easytpl.NewInitialized()
	r.LoadFiles("testdata/hello.tpl")

	err := r.Render(bf, "testdata/hello", "tom")
	art.Nil(err)
	art.Equal("hello tom", bf.String())

	art.Panics(func() {
		r.LoadFiles("not-exist.tpl")
	})
}

func TestRenderer_String(t *testing.T) {
	art := assert.New(t)
	r := easytpl.NewRenderer()
	r.MustInit()

	bf := new(bytes.Buffer)

	err := r.String(bf, `hello {{.}}`, "tom")
	art.Nil(err)
	art.Equal("hello tom", bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{.name}}`, easytpl.M{"name": "tom"})
	art.Nil(err)
	art.Equal("hello tom", bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{.Name}}`, struct {
		Name string
	}{"tom"})
	art.Nil(err)
	art.Equal("hello tom", bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{. | upper}}`, "tom")
	art.Nil(err)
	art.Equal("hello TOM", bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{. | lower}}`, "TOM")
	art.Nil(err)
	art.Equal("hello tom", bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{. | ucFirst}}`, "tom")
	art.Nil(err)
	art.Equal("hello Tom", bf.String())

	bf.Reset()
	err = r.String(bf, `hello {{. | ucFirst}}`, "")
	art.Nil(err)
	art.Equal("hello ", bf.String())

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
	r := easytpl.NewRenderer(func(r *easytpl.Renderer) {
		// r.Debug = true
		r.Layout = "layout"
	})
	art.NoError(r.Initialize())

	r.LoadStrings(map[string]string{
		"layout":       `{{ include "header" . }}, at layout, {{ yield }}, {{ include "footer" }}`,
		"admin/layout": `{{ include "header" }}, at admin layout, {{ yield }}, {{ include "footer" }}`,
		"header":       `is header:{{.}}`,
		"footer":       `is footer:{{.}}`,
		"home":         `hello {{.name}}`,
		"admin/home":   `1 at {{current}}:{{.}}`,
		"admin/login":  `2 at {{current}}:{{ . }}`,
		"other":        `at {{current}}:{{ include "not-exist" }}`,
	})

	r.LoadString("admin/reg", `main: hello {{.}}`)

	ldNames := r.TemplateNames()
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
	err = r.Partial(bf, "home", easytpl.M{"name": "tom"})
	art.Nil(err)
	art.Equal("hello tom", bf.String())

	// include not exist
	bf.Reset()
	err = r.Partial(bf, "other", easytpl.M{"name": "tom"})
	art.Nil(err)
	art.Equal("at other:", bf.String())
}

func TestRenderer_Partial(t *testing.T) {
	bf := new(bytes.Buffer)
	art := assert.New(t)

	r := easytpl.NewInitialized()

	err := r.Partial(bf, "not-exist", nil)
	art.Error(err)
}

func _TestUseExtends(t *testing.T) {
	bf := new(bytes.Buffer)
	is := assert.New(t)

	r := easytpl.NewInitialized(func(r *easytpl.Renderer) {
		r.Debug = true
		// r.ViewsDir = "testdata/extends"
	})
	r.LoadStrings(map[string]string{
		"home": `{{ extends "layout" . }}
{{ define "body" }} hello {{.}}{{ end }}`,
		// layout file
		"layout": `{{ block "header" . }}header{{ end }}
{{ block "body" . }}default{{ end }}
{{ block "footer" . }}footer{{ end }}`,
	})

	err := r.Execute(bf, "layout", "inhere")
	is.Nil(err)
	is.Equal("header\n hello inhere\nfooter", bf.String())
	bf.Reset()

	err = r.Execute(bf, "home", "inhere")
	is.Nil(err)
	is.Equal("header\n hello \nfooter\n", bf.String())
}

func _Example_Extends() {
	r := easytpl.NewInitialized()
	// load templates
	r.LoadStrings(map[string]string{
		// layout template file
		"layout.tpl": `{{ block "header" . }}header{{ end }}
{{ block "body" . }}default{{ end }}
{{ block "footer" . }}footer{{ end }}`,
		// current page template
		"home.tpl": `{{ extends "layout.tpl" . }}
{{ define "body" }}hello {{.}}{{ end }}`,
	})

	fmt.Println("- render 'layout.tpl'")
	err := r.Execute(os.Stdout, "layout", "inhere")
	if err != nil {
		panic(err)
	}

	fmt.Println("\n- render 'home.tpl'")
	err = r.Execute(os.Stdout, "home", "inhere")
	if err != nil {
		panic(err)
	}

	// Output:
	// - render 'layout.tpl'
	// header
	// default
	// footer
	// - render 'home.tpl'
	// header
	// default
	// footer
}
