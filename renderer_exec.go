package easytpl

import (
	"fmt"
	"html/template"
	"io"
)

/*************************************************************
 * render templates
 *************************************************************/

// Render a template name/file and write to the Writer.
//
// Usage:
//
//	renderer := easytpl.NewRenderer()
//	// ... ...
//	// will apply global layout setting
//	renderer.Render(http.ResponseWriter, "user/login", data)
//
//	// apply custom layout file
//	renderer.Render(http.ResponseWriter, "user/login", data, "custom-layout")
//
//	// will disable apply layout render
//	renderer.Render(http.ResponseWriter, "user/login", data, "")
func (r *Renderer) Render(w io.Writer, tplName string, v any, layouts ...string) error {
	// Apply layout render
	if layout := r.getLayoutName(layouts); layout != "" {
		r.addLayoutFuncs(layout, tplName, v)
		tplName = layout
	}

	return r.Execute(w, tplName, v)
}

// Partial is alias of the Execute()
func (r *Renderer) Partial(w io.Writer, tplName string, v any) error {
	return r.Execute(w, tplName, v)
}

// Execute render partial, will not render layout file
func (r *Renderer) Execute(w io.Writer, tplName string, v any) error {
	if !r.initialized {
		panicErr(fmt.Errorf("please call Initialize() before execute render template"))
	}

	// Render template content
	str, err := r.executeByName(tplName, v)
	if err == nil {
		_, err = w.Write([]byte(str))
	}
	return err
}

// String render a template string
func (r *Renderer) String(w io.Writer, tplString string, v any) error {
	t := template.New("string-tpl").Delims(r.Delims.Left, r.Delims.Right).Funcs(r.FuncMap)

	template.Must(t.Parse(tplString))
	return t.Execute(w, v)
}

// execute data render by template instance
func (r *Renderer) executeTemplate(tpl *template.Template, v any) (string, error) {
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

// execute data render by template name
func (r *Renderer) executeByName(name string, v any) (string, error) {
	name = r.cleanExt(name)

	// Find template instance by name
	tpl := r.templates.Lookup(name)
	if tpl == nil {
		return "", fmt.Errorf("easytpl: the template [%s] is not found", name)
	}

	return r.executeTemplate(tpl, v)
}
