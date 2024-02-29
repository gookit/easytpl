# EasyTpl

[![GoDoc](https://pkg.go.dev/badge/github.com/gookit/easytpl.svg)](https://pkg.go.dev/github.com/gookit/easytpl)
[![Coverage Status](https://coveralls.io/repos/github/gookit/easytpl/badge.svg?branch=master)](https://coveralls.io/github/gookit/easytpl?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/easytpl)](https://goreportcard.com/report/github.com/gookit/easytpl)
[![Unit-Tests](https://github.com/gookit/easytpl/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/easytpl/actions)

A simple template renderer based on the Go `html/template`, but much simpler to use. 
Support layout rendering, including templates.

> **[中文说明](README.zh-CN.md)**

## Features

- simple to use
- support loading multiple directories, multiple files
- support rendering string templates, etc.
- support layout render. 
  - eg `{{ include "header" }} {{ yield }} {{ include "footer" }}`
- support include other templates. eg `{{ include "other" }}`
- support `extends` base templates. eg `{{ extends "base.tpl" }}`
- support custom template functions
- built-in some helper methods `row`, `lower`, `upper`, `join` ...

## Godoc

- [godoc for github](https://pkg.go.dev/github.com/gookit/easytpl)

## Quick Start

```go
package main

import (
	"bytes"
	"fmt"
	
	"github.com/gookit/easytpl"
)

func main()  {
	// equals to call: easytpl.NewRenderer() + r.MustInit()
	r := easytpl.NewInited(func(r *easytpl.Renderer) {
		// setting default layout
		r.Layout = "layout" // equals to "layout.tpl"
		// templates dir. will auto load on init.
		r.ViewsDir = "testdata"
		// add template function
		r.AddFunc("myFunc", func() string {
			return "my-func"
		})
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
	
	// load named string template 
	r.LoadString("my-page", "welcome {{.}}")
	// now, you can use "my-page" as an template name
	r.Partial(bf, "my-page", "tom") // welcome tom
	bf.Reset()
	
	// more ways for load templates
	r.LoadByGlob("some/path/*", "some/path")
	r.LoadFiles("path/file1.tpl", "path/file2.tpl")
}
```

> more APIs please [GoDoc](https://pkg.go.dev/github.com/gookit/easytpl) 

## Layout Example

basic layout structure:

```text
{{ include "part0" }}{{ yield }}{{ include "part1" }}
```

> current template will render at `{{ yield }}`

example files:

```text
templates/
  |_ layouts/
  |    |_ default.tpl
  |    |_ header.tpl
  |    |_ footer.tpl
  |_ home.tpl
  |_ about.tpl
```

- **layout:** `templates/layouts/default.tpl`

```html
<html>
  <head>
    <title>layout example</title>
  </head>
  <body>
    <!-- include "layouts/header.tpl" -->
    {{ include "header" }}
    <!-- Render the current template here -->
    {{ yield }}
    <!-- include "layouts/footer.tpl" -->
    {{ include "footer" }}
  </body>
</html>
```

- `templates/layouts/header.tpl`

```html
<header>
    <h2>page header</h2>
</header>
```

- `templates/layouts/footer.tpl`

```html
<footer>
    <h2>page footer</h2>
</footer>
```

- `templates/home.tpl`

```gotemplate title="home.tpl"
  <h1>Hello, {{ .Name | upper }}</h1>
  <h2>At template {{ current_tpl }}</h2>
  <p>Lorem ipsum dolor sit amet, consectetur adipisicing elit.</p>
```

### Usage

```go
v := easytpl.NewInited(func(r *easytpl.Renderer) {
    // setting default layout
    r.Layout = "layouts/default" // equals to "layouts/default.tpl"
    // templates dir. will auto load on init.
    r.ViewsDir = "templates"
})

http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	v.Render(w, "home", easytpl.M{"Name": "tom"})
})
slog.Println("Listening port: 9100")
http.ListenAndServe(":9100", nil)
```

## `extends` example

A base template can be inherited using the `{{ extends "base.tpl" }}` statement.

> Note: The `extends` statement must be on the first line of the template file

```text
templates/
  |_ base.tpl
  |_ home.tpl
  |_ about.tpl
```

`templates/base.tpl` base template contents:

```gotemplate title="base.tpl"
<html lang="en">
  <head>
    <title>layout example</title>
  </head>
  <body>
    {{ block "content" . }}
    <h1>Hello, at base template</h1>
    {{ end }}
  </body>
</html>
```

`templates/home.tpl` contents:

```gotemplate title="home.tpl"
{{ extends "base" }}

{{ define "content" }}
  <h1>Hello, {{ .Name | upper }}</h1>
  <h2>At template {{ current_tpl }}</h2>
  <p>Lorem ipsum dolor sit amet, consectetur adipisicing elit.</p>
{{ end }}
```

### Usage

```go
package main

import (
    "net/http"
    
    "github.com/gookit/easytpl"
    "github.com/gookit/slog"
)

func main() {
    v := easytpl.NewExtends(easytpl.WithTplDirs("templates"))

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        v.Render(w, "home", easytpl.M{"Name": "tom"})
    })
    
    slog.Info("Listening port: 9100")
    http.ListenAndServe(":9100", nil)
}
```

## Available Options

```go
// Debug setting
Debug bool
// Layout template name
Layout string
// Delims define for template
Delims TplDelims
// ViewsDir the default views directory, multi use "," split
ViewsDir string
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

- method 1

```go
r := easytpl.NewRenderer()
r.Layout = "layouts/default"
// ... ...
r.MustInit()
```

- method 2

```go
r := easytpl.NewRenderer(func (r *Renderer) {
	r.Layout = "layouts/default"
	// ... ...
})
r.MustInit()
```

- method 3

```go
r := easytpl.NewInited(func (r *Renderer) {
	r.Layout = "layouts/default" 
	// ... ...
})
```

## Reference

- https://github.com/unrolled/render
- https://github.com/thedevsaddam/renderer

## License

**MIT**
