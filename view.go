/*
Package view is a simple view renderer based on the `html/template`, but much simpler to use.

Source code and other details for the project are available at GitHub:

	https://github.com/gookit/view

Usage please see example and README.
*/
package view

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// DefaultExt name
const DefaultExt = ".tpl"

// M a short type for map[string]interface{}
type M map[string]interface{}

// TplDelims for html template
type TplDelims struct {
	Left  string
	Right string
}

// create an default instance
var std = NewRenderer()

// Revert the default instance
func Revert() {
	std = NewRenderer()
}

// Default get default instance
func Default() *Renderer {
	return std
}

/*************************************************************
 * internal methods
 *************************************************************/

func (r *Renderer) compileTemplates() error {
	// Create root template engine
	r.ensureTemplates()

	dir := r.ViewsDir
	if dir == "" {
		return nil
	}

	r.debugf("will compile templates from the views dir: %s", dir)

	// Walk the supplied directory and compile any files that match our extension list.
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// Skip dir
		if info == nil || info.IsDir() {
			return nil
		}

		// Path is full path, rel is relative path
		// eg. path: "testdata/admin/footer.tpl" -> rel: "admin/footer.tpl"
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// Skip no ext
		if strings.IndexByte(rel, '.') == -1 {
			return nil
		}

		ext := filepath.Ext(rel)
		// It is valid ext
		if _, has := r.extMap[ext]; has {
			name := rel[0 : len(rel)-len(ext)]
			// create new template in the templates
			r.loadTemplateFile(name, path)
		}

		return err
	})
}

func (r *Renderer) ensureTemplates() {
	if r.templates == nil {
		r.templates = template.New("ROOT")
	}
}

func (r *Renderer) newTemplate(name string) *template.Template {
	// Create new template in the templates
	// Set delimiters and add func map
	return r.templates.New(name).
		Funcs(r.FuncMap).
		Delims(r.Delims.Left, r.Delims.Right)
}

func (r *Renderer) loadTemplateFile(tplName, file string) {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		panicErr(err)
	}

	r.fileMap[tplName] = file
	r.debugf("load template file: %s, template name is: %s", file, tplName)

	// Create new template in the templates
	template.Must(r.newTemplate(tplName).Parse(string(bs)))
}

// execute data render by template name
func (r *Renderer) executeByName(name string, v interface{}) (string, error) {
	name = r.cleanExt(name)

	// Find template instance by name
	tpl := r.templates.Lookup(name)
	if tpl == nil {
		return "", fmt.Errorf("view renderer: the template [%s] is not found", name)
	}

	return r.executeTemplate(tpl, v)
}

// execute data render by template instance
func (r *Renderer) executeTemplate(tpl *template.Template, v interface{}) (string, error) {
	// Get a buffer from the pool to write to.
	buf := r.bufPool.get()
	name := tpl.Name()

	// Current template name
	tpl.Funcs(template.FuncMap{
		"current": func() string {
			return name
		},
	})

	r.debugf("render the template: %s, override set func: current", name)
	err := tpl.Execute(buf, v)
	str := buf.String()

	// Return buffer to pool
	r.bufPool.put(buf)
	return str, err
}

// name the template name
func (r *Renderer) addLayoutFuncs(layout, name string, data interface{}) {
	tpl := r.Template(layout)
	if tpl == nil {
		panicErr(fmt.Errorf("the layout template: %s is not found, want render: %s", layout, name))
	}

	// includeHandler := func(tplName string) (template.HTML, error) {
	// 	if r.templates.Lookup(tplName) != nil {
	// 		str, err := r.executeTemplate(tplName, data)
	// 		// Return safe HTML here since we are rendering our own template.
	// 		return template.HTML(str), err
	// 	}
	//
	// 	return "", nil
	// }

	r.debugf("add funcs[yield, partial] to layout template: %s, target template: %s", layout, name)
	funcMap := template.FuncMap{
		"yield": func() (template.HTML, error) {
			str, err := r.executeByName(name, data)
			return template.HTML(str), err
		},
		// Will add data to included template
		// "include": includeHandler,
		// "partial": includeHandler,
	}

	tpl.Funcs(funcMap)
}
