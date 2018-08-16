package main

import (
	"github.com/gookit/view"
	"fmt"
	"bytes"
)

func main()  {
	r := view.NewRenderer(func(r *view.Renderer) {
		r.ViewsDir = "testdata"
	})

	r.MustInitialize()

	fmt.Println(r.LoadedNames(true))

	bf := new(bytes.Buffer)
	r.String(bf, `hello {{.}}`, "tom")
	fmt.Println(bf.String())
}