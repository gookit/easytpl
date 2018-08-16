package main

import (
	"github.com/gookit/view"
	"fmt"
	"bytes"
)

func main()  {
	// equals to call: view.NewRenderer() + r.MustInitialize()
	r := view.NewInitialized(func(r *view.Renderer) {
		r.ViewsDir = "testdata"
		r.Layout = "layout" // equals to "layout.tpl"
	})

	fmt.Println(r.LoadedNames(true))

	bf := new(bytes.Buffer)
	r.String(bf, `hello {{.}}`, "tom")
	fmt.Println(bf.String())

	// will render with layout
	r.Render(bf, "home", "tom")

	// will render with custom layout
	r.Render(bf, "home", "tom", "site/layout")
}