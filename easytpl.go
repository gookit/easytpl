/*
Package easytpl is a simple template renderer based on the `html/template`, but much simpler to use.

Source code and other details for the project are available at GitHub:

	https://github.com/gookit/easytpl

Usage please see example and README.
*/
package easytpl

import (
	"html/template"
	"io"
)

// DefaultExt name
const DefaultExt = ".tpl"

// M a short type for map[string]any
type M map[string]any

// TplDelims for html template
type TplDelims struct {
	Left  string
	Right string
}

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
func Initialize(fns ...ConfigFn) {
	std.WithConfig(fns...).MustInit()
}

/*************************************************************
 * render config func
 *************************************************************/

// WithDebug set enable debug mode.
func WithDebug(r *Renderer) { r.Debug = true }

// WithLayout set the layout template name.
func WithLayout(layoutName string) ConfigFn {
	return func(r *Renderer) {
		r.Layout = layoutName
		r.DisableLayout = false
	}
}

// DisableLayout disable the layout template.
func DisableLayout(r *Renderer) {
	r.Layout = ""
	r.DisableLayout = true
}

// WithTplDirs set template dirs
func WithTplDirs(dirs string) ConfigFn {
	return func(r *Renderer) { r.ViewsDir = dirs }
}

// WithViewDirs set template dirs, alias of WithTplDirs()
func WithViewDirs(dirs string) ConfigFn { return WithTplDirs(dirs) }

/*************************************************************
 * render templates
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
