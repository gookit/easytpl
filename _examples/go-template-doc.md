# go template doc 

> from: https://godoc.org/text/template

More Refer:

- https://liuliqiang.info/post/template-in-golang
- https://www.cnblogs.com/jkko123/p/7018406.html

## Text and spaces

> 以 `{{- `  和 ` -}}` 做开始和结尾将会自动清理对应位置的空白（注意，不是变量的空格）

For instance, when executing the template whose source is

    "{{23 -}} < {{- 45}}"

the generated output would be

    "23<45"

## Actions

### comment

```text
{{/* a comment */}}
{{- /* a comment with white space trimmed from preceding and following text */ -}}
```

A comment; discarded. May contain newlines. 
Comments do not nest and must start and end at the delimiters, as shown here.

### pipeline

```text
{{pipeline}}
```

he defaults textual representation (the same as would be printed by fmt.Print) of 
the value of the pipeline is copied to the output.

    {{"output" | printf "%q"}}

### if

```text
{{if pipeline}} T1 {{end}}
```

If the value of the pipeline is empty, no output is generated;
otherwise, T1 is executed. The empty values are false, 0, any
nil pointer or interface value, and any array, slice, map, or
string of length zero. Dot is unaffected.

### if-else

```text
{{if pipeline}} T1 {{else}} T0 {{end}}
```

If the value of the pipeline is empty, T0 is executed; otherwise, T1 is executed. Dot is unaffected.

### if-elseif

```text
{{if pipeline}} T1 {{else if pipeline}} T0 {{end}}
```

To simplify the appearance of if-else chains, the else action
of an if may include another if directly; the effect is exactly
the same as writing

	{{if pipeline}} T1 {{else}}{{if pipeline}} T0 {{end}}{{end}}

### range

    {{range pipeline}} T1 {{end}}
	
The value of the pipeline must be an array, slice, map, or channel.
If the value of the pipeline has length zero, nothing is output;
otherwise, dot is set to the successive elements of the array,
slice, or map and T1 is executed. If the value is a map and the
keys are of basic type with a defined order ("comparable"), the
elements will be visited in sorted key order.

### range-else

    {{range pipeline}} T1 {{else}} T0 {{end}}

The value of the pipeline must be an array, slice, map, or channel.
If the value of the pipeline has length zero, dot is unaffected and
T0 is executed; otherwise, dot is set to the successive elements
of the array, slice, or map and T1 is executed.

### template

    {{template "name"}}

The template with the specified name is executed with nil data.

### template-pipeline

    {{template "name" pipeline}}

The template with the specified name is executed with dot set to the value of the pipeline.

### block

    {{block "name" pipeline}} T1 {{end}}

A block is shorthand for defining a template

    {{define "name"}} T1 {{end}}
    
and then executing it in place

    {{template "name" pipeline}}

The typical use is to define a set of root templates that are
then customized by redefining the block templates within.

### with

    {{with pipeline}} T1 {{end}}

If the value of the pipeline is empty, no output is generated;
otherwise, dot is set to the value of the pipeline and T1 is executed.

### with-else

    {{with pipeline}} T1 {{else}} T0 {{end}}

If the value of the pipeline is empty, dot is unaffected and T0
is executed; otherwise, dot is set to the value of the pipeline
and T1 is executed.

## Nested template definitions

When parsing a template, another template may be defined and associated with the template being parsed. Template definitions must appear at the top level of the template, much like global variables in a Go program.

The syntax of such definitions is to surround each template declaration with a "define" and "end" action.

The define action names the template being created by providing a string constant. Here is a simple example:

```text
{{define "T1"}}ONE{{end}}
{{define "T2"}}TWO{{end}}
{{define "T3"}}{{template "T1"}} {{template "T2"}}{{end}}

{{template "T3"}}
```

This defines two templates, T1 and T2, and a third T3 that invokes the other two when it is executed. Finally it invokes T3. If executed this template will produce the text

    ONE TWO

## Examples

- block

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
