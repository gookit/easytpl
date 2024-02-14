package easytpl

import (
	"fmt"
	"html/template"
	"io"
)

/*************************************************************
 * render root
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
func (r *Renderer) Render(w io.Writer, tplName string, v any, layout ...string) error {
	// Apply layout render
	if layoutName := r.getLayoutName(layout); layoutName != "" {
		r.addLayoutFuncs(layoutName, tplName, v)
		tplName = layoutName
	}

	return r.Execute(w, tplName, v)
}

// Partial is alias of the Execute()
func (r *Renderer) Partial(w io.Writer, tplName string, v any) error {
	return r.Execute(w, tplName, v)
}

// Execute render partial, will not render layout file
func (r *Renderer) Execute(w io.Writer, tplName string, v any) error {
	if !r.init {
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
func (r *Renderer) String(w io.Writer, tplText string, v any) error {
	// must create a new template instance
	t := template.New("string-tpl").Delims(r.Delims.Left, r.Delims.Right).Funcs(r.FuncMap)

	return template.Must(t.Parse(tplText)).Execute(w, v)
}

// execute data render by template instance
func (r *Renderer) executeTemplate(tpl *template.Template, v any) (string, error) {
	// get a buffer from the pool to write to.
	buf := r.bufPool.get()
	name := tpl.Name()

	// Current template name
	tpl.Funcs(template.FuncMap{
		"current_tpl": func() string {
			return name
		},
	})

	r.debugf("render the template %q, override set func: current_tpl", name)
	err := tpl.Execute(buf, v)
	str := buf.String()

	// release buffer to pool
	r.bufPool.put(buf)
	return str, err
}

// execute data render by template name
func (r *Renderer) executeByName(name string, v any) (string, error) {
	name = r.cleanExt(name)

	// Find template instance by name
	tpl := r.root.Lookup(name)
	if tpl == nil {
		return "", fmt.Errorf("easytpl: the template [%s] is not found", name)
	}

	return r.executeTemplate(tpl, v)
}
