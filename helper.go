package view

import (
	"fmt"
	"html/template"
	"strings"
)

var globalFuncMap = template.FuncMap{
	// don't escape content
	"raw": func(s string) template.HTML {
		return template.HTML(s)
	},
	"trim":  strings.TrimSpace,
	"join":  strings.Join,
	"lower": strings.ToLower,
	"upper": strings.ToUpper,
	// uppercase first char
	"ucFirst": func(s string) string {
		if len(s) != 0 {
			f := s[0]
			// is lower
			if f >= 'a' && f <= 'z' {
				return strings.ToUpper(string(f)) + string(s[1:])
			}
		}

		return s
	},
	"yield": func() (string, error) {
		return "", fmt.Errorf("yield called with no layout defined")
	},
	// add a empty func for compile
	"current": func() string {
		return ""
	},
}

// LoadedTemplates returns loaded template instances, including ROOT itself.
func (r *Renderer) LoadedTemplates() []*template.Template {
	return r.templates.Templates()
}

// TemplateFiles returns loaded template files
func (r *Renderer) TemplateFiles() map[string]string {
	return r.fileMap
}

// TemplateNames returns loaded template names
func (r *Renderer) TemplateNames(format ...bool) string {
	str := r.templates.DefinedTemplates()
	if len(format) != 1 || format[0] == false {
		return str
	}

	str = strings.TrimLeft(str, "; ")
	return strings.NewReplacer(":", ":\n", ",", "\n").Replace(str)
}

// Templates returns root template instance
func (r *Renderer) Templates() *template.Template {
	return r.templates
}

// Template get template instance by name
func (r *Renderer) Template(name string) *template.Template {
	return r.templates.Lookup(r.cleanExt(name))
}

// cleanExt will clean file ext
// eg
// 		"some.tpl" -> "some"
// 		"path/some.tpl" -> "path/some"
func (r *Renderer) cleanExt(name string) string {
	if len(r.ExtNames) == 0 {
		return name
	}

	// has ext
	if pos := strings.LastIndexByte(name, '.'); pos > 0 {
		ext := name[pos:]
		if r.IsValidExt(ext) {
			return name[0:pos]
		}
	}

	return name
}

// IsValidExt check is valid ext name
func (r *Renderer) IsValidExt(ext string) bool {
	_, ok := r.extMap[ext]
	return ok
}

func (r *Renderer) getLayoutName(settings []string) string {
	var layout string

	disableLayout := r.DisableLayout
	if len(settings) > 0 {
		layout = strings.TrimSpace(settings[0])
		if layout == "" {
			disableLayout = true
		}
	} else {
		layout = r.Layout
	}

	// apply layout
	if !disableLayout && layout != "" {
		return layout
	}

	return ""
}

func (r *Renderer) debugf(format string, args ...interface{}) {
	if r.Debug {
		fmt.Printf("view: [DEBUG] "+format+"\n", args...)
	}
}

func panicErr(err error) {
	if err != nil {
		panic("view: [ERROR] " + err.Error())
	}
}
