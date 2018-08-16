# View renderer

[![GoDoc](https://godoc.org/github.com/gookit/view?status.svg)](https://godoc.org/github.com/gookit/view)
[![Build Status](https://travis-ci.org/gookit/view.svg?branch=master)](https://travis-ci.org/gookit/view)
[![Coverage Status](https://coveralls.io/repos/github/gookit/view/badge.svg?branch=master)](https://coveralls.io/github/gookit/view?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/view)](https://goreportcard.com/report/github.com/gookit/view)

## Features

- support layout render. 
  - eg `{{ include "header" }} {{ yield }} {{ include "footer" }}`
- support include other templates. eg `{{ include "other" }}`
- built-in some helper methods `row`, `lower`, `upper`, `join` ...

## Godoc

- [godoc for gopkg](https://godoc.org/gopkg.in/gookit/view.v1)
- [godoc for github](https://godoc.org/github.com/gookit/view)

## Usage

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
r := NewRenderer()
r.Layout = "layouts/default"
// ... ...
r.MustInitialize()
```

```go
r := NewRenderer(func (r *Renderer) {
	r.Layout = "layouts/default"
	// ... ...
})
r.MustInitialize()
```

## Reference

- https://github.com/unrolled/render
- https://github.com/thedevsaddam/renderer

## License

**MIT**
