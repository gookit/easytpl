package easytpl

import (
	"html/template"
	"io"
)

// create an default instance
var std = NewRenderer()

// Reset the default instance
func Reset() { std = NewRenderer() }

// Revert the default instance, alias of Reset()
func Revert() { Reset() }

// Default get default instance
func Default() *Renderer { return std }

// AddFunc add template func
func AddFunc(name string, fn any) { std.AddFunc(name, fn) }

// AddFuncMap add template func map
func AddFuncMap(fm template.FuncMap) { std.AddFuncMap(fm) }

// LoadString load named template string.
func LoadString(tplName string, tplString string) { std.LoadString(tplName, tplString) }

// LoadStrings load multi named template strings
func LoadStrings(sMap map[string]string) { std.LoadStrings(sMap) }

// LoadFiles load custom template files.
func LoadFiles(files ...string) { std.LoadFiles(files...) }

// LoadByGlob load templates by glob pattern.
func LoadByGlob(pattern string, baseDirs ...string) { std.LoadByGlob(pattern, baseDirs...) }

// Initialize the default instance with config func
func Initialize(fns ...OptionFn) {
	std.WithOptions(fns...).MustInit()
}

/*************************************************************
 * render templates use default instance
 *************************************************************/

// Render a template name/file with layout, write result to the Writer.
func Render(w io.Writer, tplName string, v any, layout ...string) error {
	return std.Render(w, tplName, v, layout...)
}

// Execute render partial, will not render layout file
func Execute(w io.Writer, tplName string, v any) error {
	return std.Execute(w, tplName, v)
}

// Partial is alias of the Execute()
func Partial(w io.Writer, tplName string, v any) error {
	return std.Execute(w, tplName, v)
}

// String render a template string
func String(w io.Writer, tplStr string, v any) error {
	return std.String(w, tplStr, v)
}
