# View renderer

[![GoDoc](https://godoc.org/github.com/gookit/view?status.svg)](https://godoc.org/github.com/gookit/view)
[![Build Status](https://travis-ci.org/gookit/view.svg?branch=master)](https://travis-ci.org/gookit/view)
[![Coverage Status](https://coveralls.io/repos/github/gookit/view/badge.svg?branch=master)](https://coveralls.io/github/gookit/view?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/view)](https://goreportcard.com/report/github.com/gookit/view)

View renderer by golang `html/template`, support layout render, include templates.

## Features

- simple to use
- support layout render. 
  - eg `{{ include "header" }} {{ yield }} {{ include "footer" }}`
- support include other templates. eg `{{ include "other" }}`
- built-in some helper methods `row`, `lower`, `upper`, `join` ...

## Godoc

- [godoc for gopkg](https://godoc.org/gopkg.in/gookit/view.v1)
- [godoc for github](https://godoc.org/github.com/gookit/view)

## Usage

```go
package main

import (
	"github.com/gookit/view"
	"fmt"
	"bytes"
)

func main()  {
	// equals to call: view.NewRenderer() + r.MustInitialize()
	r := view.NewInitialized(func(r *view.Renderer) {
		// setting default layout
		r.Layout = "layout" // equals to "layout.tpl"
		// templates dir. will auto load on init.
		r.ViewsDir = "testdata"
	})

	// fmt.Println(r.TemplateNames(true))

	bf := new(bytes.Buffer)

	// render template string
	r.String(bf, `hello {{.}}`, "tom")
	fmt.Print(bf.String()) // hello tom

	// render template without layout
	r.Partial(bf, "home", "tom")
	bf.Reset()

	// render with default layout
	r.Render(bf, "home", "tom")
	bf.Reset()

	// render with custom layout
	r.Render(bf, "home", "tom", "site/layout")
	bf.Reset()
	
	// load named template by string
	r.LoadString("my-page", "welcome {{.}}")
	// now, you can use "my-page" as an template name
	r.Partial(bf, "my-page", "tom") // welcome tom
	bf.Reset()
	
	// more ways for load templates
	r.LoadByGlob("some/path/*", "some/path")
	r.LoadFiles("path/file1.tpl", "path/file2.tpl")
}
```

> more API please [GoDoc](https://godoc.org/github.com/gookit/view) 

## Options

```go
// ViewsDir the default views directory
ViewsDir string
// Layout template name
Layout string
// Delims define for template
Delims TplDelims
// ExtNames allowed template extensions. eg {"tpl", "html"}
ExtNames []string
// FuncMap func map for template
FuncMap template.FuncMap
// DisableLayout disable layout. default is False
DisableLayout bool
// AutoSearchFile auto search template file, when not found on compiled templates. default is False
AutoSearchFile bool
```

### Apply options

```go
// method 1
r := NewRenderer()
r.Layout = "layouts/default"
// ... ...
r.MustInitialize()

// method 2
r := NewRenderer(func (r *Renderer) {
	r.Layout = "layouts/default"
	// ... ...
})
r.MustInitialize()

// method 3
r := NewInitialized(func (r *Renderer) {
	r.Layout = "layouts/default" 
	// ... ...
})
```

## Reference

- https://github.com/unrolled/render
- https://github.com/thedevsaddam/renderer

## License

**MIT**
