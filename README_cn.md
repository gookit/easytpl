# View renderer

[![GoDoc](https://godoc.org/github.com/gookit/view?status.svg)](https://godoc.org/github.com/gookit/view)
[![Build Status](https://travis-ci.org/gookit/view.svg?branch=master)](https://travis-ci.org/gookit/view)
[![Coverage Status](https://coveralls.io/repos/github/gookit/view/badge.svg?branch=master)](https://coveralls.io/github/gookit/view?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/view)](https://goreportcard.com/report/github.com/gookit/view)

一个简单的视图渲染器，基于golang `html/template` 封装，但更加简单易用。支持布局渲染，引入其他模板。

> **[EN README](README.md)**

- 简单更易使用
- 支持布局渲染
  - eg `{{ include "header" }} {{ yield }} {{ include "footer" }}`
- 支持引入其他模板 eg `{{ include "other" }}`
- 内置一些常用的模板方法 `row`, `lower`, `upper`, `join` ...

## Godoc

- [godoc for gopkg](https://godoc.org/gopkg.in/gookit/view.v1)
- [godoc for github](https://godoc.org/github.com/gookit/view)

## 快速开始

```go
package main

import (
	"github.com/gookit/view"
	"fmt"
	"bytes"
)

func main()  {
	// NewInitialized() 等同于同时调用: view.NewRenderer() + r.MustInitialize()
	r := view.NewInitialized(func(r *view.Renderer) {
		// 设置默认布局模板
		r.Layout = "layout" // 等同于 "layout.tpl"
		// 模板目录。将在初始化是自动加载里面的模板文件
		r.ViewsDir = "testdata"
		// 添加模板函数
		r.AddFunc("myFunc", func() string {
			return "my-func"
		})
	})

	// 输出所有载入的模板名称
	// fmt.Println(r.TemplateNames(true))

	bf := new(bytes.Buffer)

	// 渲染模板字符串
	r.String(bf, `hello {{.}}`, "tom")
	fmt.Print(bf.String()) // hello tom

	// 渲染模板，没有使用布局
	r.Partial(bf, "home", "tom")
	bf.Reset()

	// 使用默认布局渲染
	r.Render(bf, "home", "tom")
	bf.Reset()

	// 使用自定义布局渲染
	r.Render(bf, "home", "tom", "site/layout")
	bf.Reset()
	
	// 加载命名的字符串模板
	r.LoadString("my-page", "welcome {{.}}")
	// now, you can use "my-page" as an template name
	r.Partial(bf, "my-page", "tom") // welcome tom
	bf.Reset()
	
	// 更多加载模板的方法
	r.LoadByGlob("some/path/*", "some/path")
	r.LoadFiles("path/file1.tpl", "path/file2.tpl")
}
```

> 跟多API请参考 [GoDoc](https://godoc.org/github.com/gookit/view) 

## 布局示例

```text
templates/
  |_ layouts/
  |    |_ default.tpl
  |    |_ header.tpl
  |    |_ footer.tpl
  |_ home.tpl
  |_ about.tpl
```

- templates/layouts/default.tpl

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

- templates/layouts/header.tpl

```html
<header>
    <h2>page header</h2>
</header>
```

- templates/layouts/footer.tpl

```html
<footer>
    <h2>page footer</h2>
</footer>
```

- templates/home.tpl

```html
  <h1>Hello, {{ .Name | upper }}</h1>
  <h2>At template {{ current }}</h2>
  <p>Lorem ipsum dolor sit amet, consectetur adipisicing elit.</p>
```

### 使用

```go
v := view.NewInitialized(func(r *view.Renderer) {
    // setting default layout
    r.Layout = "layouts/default" // equals to "layouts/default.tpl"
    // templates dir. will auto load on init.
    r.ViewsDir = "templates"
})

http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	v.Render(w, "home", view.M{"Name": "tom"})
})
log.Println("Listening port: 9100")
http.ListenAndServe(":9100", nil)
```

## 可用选项

```go
// Debug setting
Debug bool
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

### 应用选项

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

## 参考

- https://github.com/unrolled/render
- https://github.com/thedevsaddam/renderer

## License

**MIT**
