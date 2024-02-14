package easytpl_test

import (
	"bytes"
	"testing"

	"github.com/gookit/easytpl"
	"github.com/gookit/goutil/testutil/assert"
)

func TestStd_instance(t *testing.T) {
	is := assert.New(t)

	// r := &Renderer{}
	easytpl.AddFunc("test", func() string { return "" })
	is.Panics(func() {
		easytpl.LoadFiles("testdata/layouts/home.tpl")
	})
	is.Panics(func() {
		easytpl.LoadByGlob("testdata/layouts/site/*.tpl", "testdata/layouts/site")
	})
	easytpl.AddFuncMap(map[string]interface{}{
		"test1": func() string { return "" },
	})

	r := easytpl.Default()
	r.Debug = true

	is.NoError(r.Initialize())

	// re-init
	is.Nil(r.Initialize())
	is.NotNil(r.Root())

	r.LoadByGlob("testdata/layouts/site/*.tpl", "testdata/layouts/site")
	is.Len(r.TemplateFiles(), 5)

	tpl := r.Template("header")
	is.NotNil(tpl)
	tpl = r.Template("header.tpl")
	is.NotNil(tpl)

	bf := new(bytes.Buffer)
	r1 := easytpl.NewRenderer()
	is.Panics(func() {
		_ = r1.Render(bf, "", nil)
	})

	// use layout
	r = easytpl.NewInited(func(r *easytpl.Renderer) {
		r.Layout = "layout"
		r.ViewsDir = "testdata/layouts/admin"
	})

	// including itself.
	is.Len(r.Templates(), 5+1)

	bf.Reset()
	err := r.Render(bf, "home.tpl", "tom")
	is.Nil(err)
	str := bf.String()
	is.Contains(str, "admin header")
	is.Contains(str, "home: hello")
	is.Contains(str, "admin footer")

	is.Panics(func() {
		_ = r.Render(bf, "home.tpl", "tom", "not-exist.tpl")
	})

	r = easytpl.NewInited(func(r *easytpl.Renderer) {
		r.Layout = "layout"
		r.ViewsDir = "testdata"
	})

	ns := r.TemplateNames(true)
	is.Contains(ns, "header")
	is.Contains(ns, "admin/header")
	is.Contains(ns, "site/header")

	easytpl.Revert() // Revert
}
