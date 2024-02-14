# go template doc 模板使用参考

Package template implements data-driven templates for generating textual output.
包模板实现用于生成文本输出的数据驱动模板。

To generate HTML output, see html/template, which has the same interface as this package but automatically secures HTML output against certain attacks.
要生成 HTML 输出，请参阅 html/template，它与此包具有相同的接口，但会自动保护 HTML 输出免受某些攻击。

The input text for a template is UTF-8-encoded text in any format. "Actions"--data evaluations or control structures--are delimited by "{{" and "}}"; all text outside actions is copied to the output unchanged.
模板的输入文本是任意格式的 UTF-8 编码文本。“操作”——数据评估或控制结构——由“{{”和“}}”分隔;动作外的所有文本都将原封不动地复制到输出中。

Once parsed, a template may be executed safely in parallel, although if parallel executions share a Writer the output may be interleaved.
解析后，模板可以安全地并行执行，但如果并行执行共享一个编写器，则输出可能是交错的。

## Text and spaces

By default, all text between actions is copied verbatim when the template is executed. For example, the string " items are made of " in the example above appears on standard output when the program is run.
默认情况下，执行模板时，将逐字复制操作之间的所有文本。例如，上面示例中的字符串“items is made of ”出现在程序运行时的标准输出上。

> 以 `{{- `  和 ` -}}` 做开始和结尾将会自动清理对应位置的空白（注意，不是变量的空格）

For instance, when executing the template whose source is
例如，当执行其源为

    "{{23 -}} < {{- 45}}"

the generated output would be
生成的输出将是

    "23<45"

For this trimming, the definition of white space characters is the same as in Go: space, horizontal tab, carriage return, and newline.
对于此修剪，空格字符的定义与 Go 中的定义相同：空格、水平制表符、回车符和换行符。

{{.}}
模板语法都包含在{{和}}中间，其中{{.}}中的点表示当前对象。

## Actions 操作

Here is the list of actions. "Arguments" and "pipelines" are evaluations of data, defined in detail in the corresponding sections that follow.
以下是操作列表。“参数”和“管道”是对数据的评估，在后面的相应部分中进行了详细定义。

```go
{{/* a comment */}}
{{- /* a comment with white space trimmed from preceding and following text */ -}}
	A comment; discarded. May contain newlines.
	Comments do not nest and must start and end at the
	delimiters, as shown here.

{{pipeline}}
	The default textual representation (the same as would be
	printed by fmt.Print) of the value of the pipeline is copied
	to the output.

{{if pipeline}} T1 {{end}}
	If the value of the pipeline is empty, no output is generated;
	otherwise, T1 is executed. The empty values are false, 0, any
	nil pointer or interface value, and any array, slice, map, or
	string of length zero.
	Dot is unaffected.

{{if pipeline}} T1 {{else}} T0 {{end}}
	If the value of the pipeline is empty, T0 is executed;
	otherwise, T1 is executed. Dot is unaffected.

{{if pipeline}} T1 {{else if pipeline}} T0 {{end}}
	To simplify the appearance of if-else chains, the else action
	of an if may include another if directly; the effect is exactly
	the same as writing
		{{if pipeline}} T1 {{else}}{{if pipeline}} T0 {{end}}{{end}}

{{range pipeline}} T1 {{end}}
	The value of the pipeline must be an array, slice, map, or channel.
	If the value of the pipeline has length zero, nothing is output;
	otherwise, dot is set to the successive elements of the array,
	slice, or map and T1 is executed. If the value is a map and the
	keys are of basic type with a defined order, the elements will be
	visited in sorted key order.

{{range pipeline}} T1 {{else}} T0 {{end}}
	The value of the pipeline must be an array, slice, map, or channel.
	If the value of the pipeline has length zero, dot is unaffected and
	T0 is executed; otherwise, dot is set to the successive elements
	of the array, slice, or map and T1 is executed.

{{break}}
	The innermost {{range pipeline}} loop is ended early, stopping the
	current iteration and bypassing all remaining iterations.
```

### arguments 参数

An argument is a simple value, denoted by one of the following.
参数是一个简单的值，由以下值之一表示。

* A boolean, string, character, integer, floating-point, imaginary or complex constant in Go syntax. These behave like Go's untyped constants. Note that, as in Go, whether a large integer constant overflows when assigned or passed to a function can depend on whether the host machine's ints are 32 or 64 bits.
* The keyword `nil`, representing an untyped Go nil.
* The character '.' (period): `.` The result is the value of dot.
* A variable name, which is a (possibly empty) alphanumeric string preceded by a dollar sign, such as `$piOver2` or `$` The result is the value of the variable. Variables are described below.
* The name of a field of the data, which must be a struct, preceded by a period, such as `.Field` The result is the value of the field. Field invocations may be chained: `.Field1.Field2` Fields can also be evaluated on variables, including chaining: `$x.Field1.Field2`
* The name of a key of the data, which must be a map, preceded by a period, such as `.Key` The result is the map element value indexed by the key. Key invocations may be chained and combined with fields to any depth: `.Field1.Key1.Field2.Key2` Although the key must be an alphanumeric identifier, unlike with field names they do not need to start with an upper case letter. Keys can also be evaluated on variables, including chaining: `$x.key1.key2`

参数的计算结果可以是任何类型的;如果它们是指针，则实现会在需要时自动定向到基类型。如果计算产生函数值（如结构的函数值字段），则不会自动调用该函数，但可以将其用作 if 操作等的真值。要调用它，请使用下面定义的调用函数。

### comment 注释

```gotemplate
{{/* a comment */}}
{{- /* a comment with white space trimmed from preceding and following text */ -}}
```

A comment; discarded. May contain newlines. 
Comments do not nest and must start and end at the delimiters, as shown here.

### pipeline 管道

A pipeline is a possibly chained sequence of "commands". A command is a simple value (argument) or a function or method call, possibly with multiple arguments:
 管道是可能链接的“命令”序列。命令是一个简单的值（参数）或函数或方法调用，可能有多个参数：

> pipeline 是指产生数据的操作

```gotemplate
Argument
	The result is the value of evaluating the argument.
.Method [Argument...]
	The method can be alone or the last element of a chain but,
	unlike methods in the middle of a chain, it can take arguments.
	The result is the value of calling the method with the
	arguments:
		dot.Method(Argument1, etc.)
functionName [Argument...]
	The result is the value of calling the function associated
	with the name:
		function(Argument1, etc.)
	Functions and function names are described below.
```

A pipeline may be "chained" by separating a sequence of commands with pipeline characters '|'. In a chained pipeline, the result of each command is passed as the last argument of the following command. The output of the final command in the pipeline is the value of the pipeline.
可以通过用管道字符“|”分隔命令序列来“链接”管道。在链接管道中，每个命令的结果作为以下命令的最后一个参数传递。管道中最后一个命令的输出是管道的值。

The output of a command will be either one value or two values, the second of which has type error. If that second value is present and evaluates to non-nil, execution terminates and the error is returned to the caller of Execute.
命令的输出将是一个值或两个值，其中第二个值具有类型错误。如果第二个值存在并且计算结果为非 `nil`，则执行终止，错误将返回给 Execute 的调用方。

```gotemplate
{{pipeline}}
{{pipeline | pipeline}}
{{pipeline | pipeline | pipeline}}
```

下面是Pipeline的几种示例，它们都输出"output"：

```gotemplate
{{`"output"`}}
{{printf "%q" "output"}}
{{"output" | printf "%q"}}
{{printf "%q" (print "out" "put")}}
{{"put" | printf "%s%s" "out" | printf "%q"}}
{{"output" | printf "%s" | printf "%q"}}
```

pipeline是指产生数据的操作。比如 `{{.}}`、`{{.Name}}`等。Go的模板语法中支持使用管道符号`|`链接多个命令，用法和unix下的管道类似：`|`前面的命令会将运算结果(或返回值)传递给后一个命令的最后一个位置。

> 注意 : 并不是只有使用了 `|` 才是pipeline。Go的模板语法中，pipeline的概念是传递数据，只要能产生数据的，都是pipeline。

### variables 变量

action里可以初始化一个变量来捕获管道的执行结果。初始化语法如下：

    {{ $variable := pipeline }}

其中 `$variable` 是变量的名字。声明变量的action不会产生任何输出。

## 基本使用

### if 条件判断

```gotemplate
{{if pipeline}} T1 {{end}}
```

if-else:

```gotemplate
{{if pipeline}} T1 {{else}} T0 {{end}}
```

if-elseif:

```gotemplate
{{if pipeline}} T1 {{else if pipeline}} T0 {{end}}
```

可以嵌套使用：

	{{if pipeline}} T1 {{else}}{{if pipeline}} T0 {{end}}{{end}}

### range 遍历

Go的模板语法中使用range关键字进行遍历，有以下两种写法，其中pipeline的值必须是数组、切片、字典或者通道。

```gotemplate
# 如果pipeline的值其长度为0，不会有任何输出
{{range pipeline}} T1 {{end}}
```

range-else:

```gotemplate
# 如果pipeline的值其长度为0，则会执行T0。
{{range pipeline}} T1 {{else}} T0 {{end}}
```

需注意的是，range的参数部分是pipeline，所以在迭代的过程中是可以进行赋值的。但有两种赋值情况：

```gotemplate
{{range $value := .}} 
    {{/* $value是当前正在迭代元素的值 */}}
{{ end }}
{{range $key,$value := .}} 
    {{/* $key是索引值，$value是当前正在迭代元素的值 */}}
{{ end }}
```

### with

with用来设置 `.` 的值。两种格式：

如果pipeline为empty不产生输出，否则将dot设为pipeline的值并执行T1。不修改外面的dot。

```gotemplate
{{with pipeline}} T1 {{end}}
```

with-else:

如果pipeline为empty，不改变dot并执行T0，否则dot设为pipeline的值并执行T1。

```gotemplate
{{with pipeline}} T1 {{else}} T0 {{end}}
```

------

## template 嵌套模板

- `define` 可以直接在待解析内容中定义一个模板，这个模板会加入到common结构组中，并关联到关联名称上
- `template` 定义了模板之后，可以使用template这个action来执行模板。

我们可以在template中嵌套其他的template。这个template可以是单独的文件，也可以是通过define定义的template。

    {{template "name"}}

The template with the specified name is executed with nil data.
将具有指定名称的模板以零数据执行。

template-pipeline:

    {{template "name" pipeline}}

The template with the specified name is executed with dot set to the value of the pipeline.
执行具有指定名称的模板，并将点设置为管道的值。

### Nested template definitions 嵌套模板定义

When parsing a template, another template may be defined and associated with the template being parsed. Template definitions must appear at the top level of the template, much like global variables in a Go program.
解析模板时，可以定义另一个模板并将其与正在解析的模板相关联。模板定义必须出现在模板的顶层，就像 Go 程序中的全局变量一样。

The syntax of such definitions is to surround each template declaration with a "define" and "end" action.
此类定义的语法是用“define”和“end”操作将每个模板声明括起来。

The define action names the template being created by providing a string constant. Here is a simple example:
define 操作通过提供字符串常量来命名正在创建的模板。下面是一个简单的示例：

```gotemplate
{{define "T1"}}ONE{{end}}
{{define "T2"}}TWO{{end}}
{{define "T3"}}{{template "T1"}} {{template "T2"}}{{end}}

{{template "T3"}}
```

> 注意，模板之间的变量是不会继承的。

This defines two templates, T1 and T2, and a third T3 that invokes the other two when it is executed. Finally it invokes T3. If executed this template will produce the text
这将定义两个模板，即 T1 和 T2，以及第三个 T3，该模板在执行时调用其他两个模板。最后，它调用 T3。如果执行，此模板将生成文本

    ONE TWO

By construction, a template may reside in only one association. If it's necessary to have a template addressable from multiple associations, the template definition must be parsed multiple times to create distinct *Template values, or must be copied with Template.Clone or Template.AddParseTree.
通过构造，模板只能驻留在一个关联中。如果需要从多个关联中寻址模板，则必须多次分析模板定义以创建不同的 *Template 值，或者必须使用 Template.Clone 或 Template.AddParseTree 进行复制。

Parse may be called multiple times to assemble the various associated templates; see ParseFiles, ParseGlob, Template.ParseFiles and Template.ParseGlob for simple ways to parse related templates stored in files.
可以多次调用 Parse 来组装各种关联的模板;请参阅 ParseFiles、ParseGlob、Template.ParseFiles 和 Template.ParseGlob，了解分析存储在文件中的相关模板的简单方法。

A template may be executed directly or through Template.ExecuteTemplate, which executes an associated template identified by name. To invoke our example above, we might write,
模板可以直接执行，也可以通过 Template.ExecuteTemplate 执行，后者执行由名称标识的关联模板。为了引用我们上面的例子，我们可以这样写，

```go
err := tmpl.Execute(os.Stdout, "no data needed")
if err != nil {
    log.Fatalf("execution failed: %s", err)
}
```

or to invoke a particular template explicitly by name,
或按名称显式调用特定模板，

```go
err := tmpl.ExecuteTemplate(os.Stdout, "T2", "no data needed")
if err != nil {
    log.Fatalf("execution failed: %s", err)
}
```

### block 块

`block` 等价于define定义一个名为name的模板，并在"有需要"的地方执行这个模板，执行时将 `.` 设置为pipeline的值。

但应该注意，block的第一个动作是执行名为name的模板，如果不存在，则在此处自动定义这个模板，并执行这个临时定义的模板。
换句话说，block可以认为是设置一个默认模板。

典型用途是定义一组根模板，然后通过重新定义其中的块模板来自定义这些根模板。

    {{block "name" pipeline}} T1 {{end}}

上面语句的含义是执行名为 `name` 的模板，如果不存在，则在此处自动定义这个模板，并执行这个临时定义的模板内容T1。

> 注意，block模板的定义是在执行时才会执行的，所以block模板的定义可以在执行之前。

`block` 块是定义模板的简写, 上面的等同于(define + template):

    {{define "name"}} T1 {{end}}
    // 然后执行它
    {{template "name" pipeline}}

### with

    {{with pipeline}} T1 {{end}}

If the value of the pipeline is empty, no output is generated;
otherwise, dot is set to the value of the pipeline and T1 is executed.

### with-else

    {{with pipeline}} T1 {{else}} T0 {{end}}

If the value of the pipeline is empty, dot is unaffected and T0
is executed; otherwise, dot is set to the value of the pipeline
and T1 is executed.

## 预定义函数

执行模板时，函数从两个函数字典中查找：首先是模板函数字典，然后是全局函数字典。一般不在模板内定义函数，而是使用Funcs方法添加函数到模板里。

预定义的全局函数如下：

```gotemplate
and
    函数返回它的第一个empty参数或者最后一个参数；
    就是说"and x y"等价于"if x then y else x"；所有参数都会执行；
or
    返回第一个非empty参数或者最后一个参数；
    亦即"or x y"等价于"if x then x else y"；所有参数都会执行；
not
    返回它的单个参数的布尔值的否定
len
    返回它的参数的整数类型长度
index
    执行结果为第一个参数以剩下的参数为索引/键指向的值；
    如"index x 1 2 3"返回x[1][2][3]的值；每个被索引的主体必须是数组、切片或者字典。
print
    即fmt.Sprint
printf
    即fmt.Sprintf
println
    即fmt.Sprintln
html
    返回其参数文本表示的HTML逸码等价表示。
urlquery
    返回其参数文本表示的可嵌入URL查询的逸码等价表示。
js
    返回其参数文本表示的JavaScript逸码等价表示。
call
    执行结果是调用第一个参数的返回值，该参数必须是函数类型，其余参数作为调用该函数的参数；
    如"call .X.Y 1 2"等价于go语言里的dot.X.Y(1, 2)；
    其中Y是函数类型的字段或者字典的值，或者其他类似情况；
    call的第一个参数的执行结果必须是函数类型的值（和预定义函数如print明显不同）；
    该函数类型值必须有1到2个返回值，如果有2个则后一个必须是error接口类型；
    如果有2个返回值的方法返回的error非nil，模板执行会中断并返回给调用模板执行者该错误；
```

### 比较函数

布尔函数会将任何类型的零值视为假，其余视为真。

下面是定义为函数的二元比较运算的集合：

    eq      如果 arg1 == arg2 则返回真
    ne      如果 arg1 != arg2 则返回真
    lt      如果 arg1 < arg2 则返回真
    le      如果 arg1 <= arg2 则返回真
    gt      如果 arg1 > arg2 则返回真
    ge      如果 arg1 >= arg2 则返回真

为了简化多参数相等检测，`eq`（只有eq）可以接受2个或更多个参数，它会将第一个参数和其余参数依次比较，返回下式的结果：

    {{eq arg1 arg2 arg3}}

> **NOTE**: 比较函数只适用于基本类型（或重定义的基本类型，如”type Celsius float32”）。但是，整数和浮点数不能互相比较。

### 自定义函数

Go的模板支持自定义函数。通过`Funcs`方法添加函数到模板里。

```go
template.New("hello").Funcs(template.FuncMap{"myfunc": func(s string) string {
    return strings.Upper(s) 
}})
```

模板中使用：

```gotemplate
{{myfunc "hello"}}
```

## Examples 使用示例

### 基础示例

```go
package main

import (
	"log"
	"os"
	"text/template"
)

func main() {
	// Define a template.
	const letter = `
Dear {{.Name}},
{{if .Attended}}
It was a pleasure to see you at the wedding.
{{- else}}
It is a shame you couldn't make it to the wedding.
{{- end}}
{{with .Gift -}}
Thank you for the lovely {{.}}.
{{end}}
Best wishes,
Josie
`

	// Prepare some data to insert into the template.
	type Recipient struct {
		Name, Gift string
		Attended   bool
	}
	var recipients = []Recipient{
		{"Aunt Mildred", "bone china tea set", true},
		{"Uncle John", "moleskin pants", false},
		{"Cousin Rodney", "", false},
	}

	// Create a new template and parse the letter into it.
	t := template.Must(template.New("letter").Parse(letter))

	// Execute the template for each recipient.
	for _, r := range recipients {
		err := t.Execute(os.Stdout, r)
		if err != nil {
			log.Println("executing template:", err)
		}
	}
}
```

### block 使用示例

```go
const (
    master  = `Names:{{block "list" .}}
{{range .}}{{println "-" .}}{{end}}{{end}}`
    overlay = `{{define "list"}} {{join . ", "}}{{end}} `
)
var (
    funcs     = template.FuncMap{"join": strings.Join}
    guardians = []string{"Gamora", "Groot", "Nebula", "Rocket", "Star-Lord"}
)
masterTmpl, err := template.New("master").Funcs(funcs).Parse(master)
if err != nil {
    log.Fatal(err)
}
overlayTmpl, err := template.Must(masterTmpl.Clone()).Parse(overlay)
if err != nil {
    log.Fatal(err)
}
if err := masterTmpl.Execute(os.Stdout, guardians); err != nil {
    log.Fatal(err)
}
if err := overlayTmpl.Execute(os.Stdout, guardians); err != nil {
    log.Fatal(err)
}
```

output:

```text
Names:
- Gamora
- Groot
- Nebula
- Rocket
- Star-Lord
Names: Gamora, Groot, Nebula, Rocket, Star-Lord
```

## 参考文档

> - https://pkg.go.dev/text/template
> - [Go标准库：深入剖析Go template](https://www.cnblogs.com/f-ck-need-u/p/10035768.html)
> - [Go template高级用法、深入详解、手册、指南、剖析](https://www.topgoer.com/%E5%B8%B8%E7%94%A8%E6%A0%87%E5%87%86%E5%BA%93/template.html)
