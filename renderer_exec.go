package easytpl

import (
	"fmt"
	"html/template"
	"io"

	"github.com/gookit/goutil/errorx"
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
func (r *Renderer) Execute(w io.Writer, tplName string, v any) (err error) {
	r.requireInit("please call Init() before execute template")

	// render template by name
	bs, err := r.executeByName(tplName, v)
	if err == nil {
		_, err = w.Write(bs)
	}
	return err
}

// String render a template string with data
func (r *Renderer) String(w io.Writer, tplText string, v any) error {
	// must create a new tmp template instance
	t := r.newTemplate("string-tpl")
	return template.Must(t.Parse(tplText)).Execute(w, v)
}

// execute data render by template name
func (r *Renderer) executeByName(name string, v any) ([]byte, error) {
	tpl := r.Template(name)
	if tpl == nil {
		return nil, fmt.Errorf("easytpl: execute template %q is not found", name)
	}
	return r.executeTemplate(tpl, v)
}

// execute data render by template instance
func (r *Renderer) executeTemplate(tpl *template.Template, v any) ([]byte, error) {
	// get a buffer from the pool to write to.
	buf := r.bufPool.get()
	// name := tpl.Name()
	name := tpl.Tree.Name

	tpl.Funcs(template.FuncMap{
		// get current template name
		"current_tpl": func() string {
			return name
		},
	})

	r.debugf("execute the template %q, override set func: current_tpl", name)
	err := tpl.Execute(buf, v)
	bts := buf.Bytes()

	// release buffer to pool
	r.bufPool.put(buf)
	return bts, err
}

func (r *Renderer) handleInclude(tplName string, data ...any) (template.HTML, error) {
	tpl := r.Template(tplName)
	if tpl == nil {
		return "", errorx.Ef("the include template %q is not found", tplName)
	}

	// do render template with data
	var v any
	if len(data) == 1 {
		v = data[0]
	}

	bs, err := r.executeTemplate(tpl, v)
	return template.HTML(bs), err
}
