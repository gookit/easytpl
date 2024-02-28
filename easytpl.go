/*
Package easytpl is a simple template renderer based on the `html/template`, but much simpler to use.

Source code and other details for the project are available at GitHub:

	https://github.com/gookit/easytpl

Usage please see example and README.
*/
package easytpl

import (
	"fmt"
	"html/template"
)

// DefaultExt name
const DefaultExt = ".tpl"
const DefaultExt1 = ".tmpl"

// M a short type for map[string]any
type M map[string]any

// TplDelims for html template
type TplDelims struct {
	Left  string
	Right string
}

// default global built in func map
var builtInFuncMap = template.FuncMap{
	// don't escape content
	"raw": func(s string) template.HTML { return template.HTML(s) },
	// add some empty func for resolve compile error
	"yield": func() (string, error) {
		return "", fmt.Errorf("yield called with no layout defined")
	},
	"current_tpl": func() string { return "" },
}

// Options for renderer
type Options struct {
	// Debug mode for development.
	Debug bool
	// Delims define for template. default is "{{", "}}"
	Delims TplDelims
	// ViewsDir the default views directory, multi dirs use "," split
	ViewsDir string
	// ExtNames supported template extensions, without dot prefix. eg {"tpl", "html"}
	ExtNames []string
	// FuncMap func map for template
	FuncMap template.FuncMap

	// Layout template name for default.
	Layout string
	// DisableLayout disable apply layout render. default is False
	DisableLayout bool

	// EnableExtends enable extends feature. default is False
	EnableExtends bool
	// ExtendsBase template file map. available when extends feature is enabled.
	// 	- Key is tpl name, value is base tpl file path.
	//
	// Example:
	// 	{"base": "/path/to/base.tpl"}
	// Use on template page.tpl:
	// 	{{ extends "base" }}
	// 	{{ define "body" }} ... {{ end }}
	ExtendsBase map[string]string
	// AutoSearchFile
	// TODO: auto search template file, when not found on compiled templates. default is False
	AutoSearchFile bool
}

// OptionFn for renderer
type OptionFn func(r *Renderer)

// New create a new template renderer, but not initialized.
func New(fns ...OptionFn) *Renderer { return NewRenderer(fns...) }

// NewInited create a new and initialized template renderer. alias of NewInitialized()
func NewInited(fns ...OptionFn) *Renderer {
	return NewRenderer(fns...).MustInit()
}

// NewExtends create a new and initialized template renderer. default enable extends feature.
func NewExtends(fns ...OptionFn) *Renderer {
	return NewRenderer(fns...).WithOptions(EnableExtends).MustInit()
}

// NewInitialized create a new and initialized view renderer.
func NewInitialized(fns ...OptionFn) *Renderer {
	return NewRenderer(fns...).MustInit()
}

/*************************************************************
 * renderer options config func
 *************************************************************/

// WithDebug set enable debug mode.
func WithDebug(r *Renderer) { r.Debug = true }

// WithLayout set the layout template name.
func WithLayout(layoutName string) OptionFn {
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

// EnableExtends enable extends feature.
func EnableExtends(r *Renderer) { r.EnableExtends = true }

// WithTplDirs set template dirs
func WithTplDirs(dirs string) OptionFn {
	return func(r *Renderer) { r.ViewsDir = dirs }
}

// WithViewDirs set template dirs, alias of WithTplDirs()
func WithViewDirs(dirs string) OptionFn { return WithTplDirs(dirs) }
