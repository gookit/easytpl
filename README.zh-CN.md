# EasyTpl

[![GoDoc](https://pkg.go.dev/github.com/gookit/easytpl?status.svg)](https://pkg.go.dev/github.com/gookit/easytpl)
[![Coverage Status](https://coveralls.io/repos/github/gookit/easytpl/badge.svg?branch=master)](https://coveralls.io/github/gookit/easytpl?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/easytpl)](https://goreportcard.com/report/github.com/gookit/easytpl)
[![Unit-Tests](https://github.com/gookit/easytpl/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/easytpl/actions)

一个简单的模板渲染器，基于Golang `html/template` 封装，但更加简单易用。

> **[EN README](README.md)**

- 简单，易使用
- 支持加载多目录，多文件
- 支持渲染字符串模板等
- 支持布局文件渲染
  - eg `{{ include "header" }} {{ yield }} {{ include "footer" }}`
- 支持引入其他模板 eg `{{ include "other" }}`
- 支持使用 `extends` 继承基础模板. eg `{{ extends "base.tpl" }}`
- 内置一些常用的模板方法 `row`, `lower`, `upper`, `join` ...

## GoDoc

- [GoDoc for GitHub](https://pkg.go.dev/github.com/gookit/easytpl)

## 快速开始

```go
package main

import (
	"bytes"
	"fmt"
	
	"github.com/gookit/easytpl"
)

func main()  {
	// NewInitialized() 等同于同时调用: easytpl.NewRenderer() + r.MustInitialize()
	r := easytpl.NewInitialized(func(r *easytpl.Renderer) {
		// 设置默认布局模板
		r.Layout = "layout" // 等同于 "layout.tpl"
		// 模板目录。将在初始化是自动加载编译里面的模板文件
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
	_= r.String(bf, `hello {{.}}`, "tom")
	fmt.Print(bf.String()) // hello tom

	// 渲染模板，没有使用布局
	_= r.Partial(bf, "home", "tom")
	bf.Reset()

	// 使用默认布局渲染
	_= r.Render(bf, "home", "tom")
	bf.Reset()

	// 使用自定义布局渲染
	_= r.Render(bf, "home", "tom", "site/layout")
	bf.Reset()
	
	// 加载命名的字符串模板
	r.LoadString("my-page", "welcome {{.}}")
	// now, you can use "my-page" as an template name
	_= r.Partial(bf, "my-page", "tom") // welcome tom
	bf.Reset()
	
	// 更多加载模板的方法
	r.LoadByGlob("some/path/*", "some/path")
	r.LoadFiles("path/file1.tpl", "path/file2.tpl")
}
```

> 跟多API请参考 [GoDoc](https://pkg.go.dev/github.com/gookit/easytpl) 

## layout 布局示例

- `include` 包含其他模板

```text
templates/
  |_ layouts/
  |    |_ default.tpl
  |    |_ header.tpl
  |    |_ footer.tpl
  |_ home.tpl
  |_ about.tpl
```

`templates/layouts/default.tpl` 布局文件:

```html
<html lang="en">
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

`templates/home.tpl, templates/about.tpl` 模板文件:

```html title="home.tpl"
  <h1>Hello, {{ .Name | upper }}</h1>
  <h2>At template {{ current_tpl }}</h2>
  <p>Lorem ipsum dolor sit amet, consectetur adipisicing elit.</p>
```

### 使用示例

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

## `extends` 继承示例

可以使用 `extends` 语句继承基础模板. eg `{{ extends "base.tpl" }}`

> 注意: `extends` 语句必须在模板文件的第一行

```text
templates/
  |_ base.tpl
  |_ home.tpl
  |_ about.tpl
```

`templates/base.tpl` 基础模板文件:

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

`templates/home.tpl, templates/about.tpl` 模板文件:

```gotemplate title="home.tpl"
  {{ extends "base" }}
{{ define "content" }}
  <h1>Hello, {{ .Name | upper }}</h1>
  <h2>At template {{ current_tpl }}</h2>
  <p>Lorem ipsum dolor sit amet, consectetur adipisicing elit.</p>
{{ end }}
```

### 使用示例

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

## 可用选项

```go
// 开启调试
Debug bool
// 默认的视图模板目录，可以是多个目录，使用逗号分隔
ViewsDir string
// 布局模板名称
Layout string
// 模板变量分隔符定义
Delims TplDelims
// 允许的模板扩展. eg {"tpl", "html"}
ExtNames []string
// 添加自定义模板函数
FuncMap template.FuncMap
// 禁用布局。默认值为False
DisableLayout bool
// TODO 如果在已编译模板上找不到，自动查找模板文件。默认值为False
AutoSearchFile bool
```

### 应用选项

- 方法 1

```go
r := easytpl.NewRenderer()
r.Layout = "layouts/default"
// ... ...
r.MustInit()
```

- 方法 2

```go
r := easytpl.NewRenderer(func (r *Renderer) {
	r.Layout = "layouts/default"
	// ... ...
})
r.MustInit()
```

- 方法 3 (简单快捷)

```go
r := easytpl.NewInited(func (r *Renderer) {
	r.Layout = "layouts/default" 
	// ... ...
})
```

## 参考

- https://github.com/unrolled/render
- https://github.com/thedevsaddam/renderer

## License

**[MIT](LICENSE)**
