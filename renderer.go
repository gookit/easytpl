package view

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"fmt"
)

const (
	DefExt = ".tpl"
)

// TplDelims for html template
type TplDelims struct {
	Left  string
	Right string
}

// Renderer definition
type Renderer struct {
	extMap  map[string]uint8
	bufPool *bufferPool
	// templates is root template instance.
	// it is like a map, contains all parsed templates
	// {
	// 	"tpl name0": *template.Template,
	// 	"tpl name1": *template.Template,
	// 	... ...
	// }
	templates *template.Template

	// ViewsDir the default views directory
	ViewsDir string
	// Layout template setting
	Layout string
	// Delims define for template
	Delims TplDelims
	// ExtNames eg {"tpl", "html"}
	ExtNames []string
	// FuncMap func map for template
	FuncMap template.FuncMap
	// DisableLayout disable layout
	DisableLayout bool
}

// NewRenderer create a default view renderer
func NewRenderer(fns ...func(r *Renderer)) *Renderer {
	r := &Renderer{
		Delims:   TplDelims{"{{", "}}"},
		FuncMap:  make(template.FuncMap),
		ExtNames: []string{"tpl", "html"},
	}

	if len(fns) == 1 {
		fns[0](r)
	}

	r.bufPool = newBufferPool()

	return r
}

/*************************************************************
 * prepare for viewRenderer
 *************************************************************/

// AddFunc add template func
func (r *Renderer) AddFunc(name string, fn interface{}) {
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		panicErr(fmt.Errorf("the template [%s] func must be a callable function", name))
	}

	r.FuncMap[name] = fn
}

// AddFuncMap add template func map
func (r *Renderer) AddFuncMap(fm template.FuncMap) {
	for name, fn := range fm {
		r.FuncMap[name] = fn
	}
}

// Initialize templates in the viewsDir, add do some prepare works.
func (r *Renderer) Initialize() error {
	if len(r.ExtNames) == 0 {
		r.ExtNames = []string{DefExt}
	}

	r.extMap = make(map[string]uint8, len(r.ExtNames))
	for _, ext := range r.ExtNames {
		if ext[0] != '.' {
			ext = "." + ext
		}

		r.extMap[ext] = 0
	}

	// add template func
	r.AddFuncMap(globalFuncMap)
	r.AddFunc("include", func(tplName string, data ...interface{}) (template.HTML, error) {
		if r.Template(tplName) != nil {
			var v interface{}
			if len(data) == 1 {
				v = data[0]
			}

			buf, err := r.executeTemplate(tplName, v)
			return template.HTML(buf.String()), err
		}

		return "", nil
	})

	// create root template engine
	r.ensureTemplates()

	// compile templates
	if err := r.compileTemplates(); err != nil {
		return err
	}

	return nil
}

// MustInitialize compile templates and report error
func (r *Renderer) MustInitialize() {
	panicErr(r.Initialize())
}

// LoadByGlob load templates by glob
// usage:
// 		r.LoadByGlob("views/*")
// 		r.LoadByGlob("views/**/*")
func (r *Renderer) LoadByGlob(pattern string) {
	r.ensureTemplates()

	files, err := filepath.Glob(pattern)
	panicErr(err)

	r.LoadFiles(files...)
}

// LoadFiles load template files
// usage:
// 		r.LoadFiles("path/file1.tpl", "path/file2.tpl")
func (r *Renderer) LoadFiles(files ...string) {
	r.ensureTemplates()

	for _, file := range files {
		ext := filepath.Ext(file)

		if r.IsValidExt(ext) {
			name := filepath.ToSlash(file[0:len(file) - len(ext)])
			r.loadTemplateFile(name, file)
		}
	}
}

// LoadString load template string
func (r *Renderer) LoadString(tplName string, tplString string) {
	r.ensureTemplates()

	// create new template in the templates
	template.Must(r.newTemplate(tplName).Parse(tplString))
}

// LoadString load template strings
func (r *Renderer) LoadStrings(sMap map[string]string) {
	for name, tplStr := range sMap {
		r.LoadString(name, tplStr)
	}
}

func (r *Renderer) compileTemplates() error {
	dir := r.ViewsDir
	if dir == "" {
		return nil
	}

	// Walk the supplied directory and compile any files that match our extension list.
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() { // skip dir
			return nil
		}

		// path is full path, rel is relative path
		// eg path: "testdata/admin/footer.tpl" -> rel: "admin/footer.tpl"
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if strings.IndexByte(rel, '.') == -1 { // skip no ext
			return nil
		}

		ext := filepath.Ext(rel)
		// is valid ext
		if _, has := r.extMap[ext]; has {
			name := rel[0 : len(rel)-len(ext)]

			// create new template in the templates
			r.loadTemplateFile(name, path)
		}

		return err
	})
}

func (r *Renderer) ensureTemplates() {
	if r.templates != nil {
		return
	}

	r.templates = template.New("ROOT").Delims(r.Delims.Left, r.Delims.Right)
}

func (r *Renderer) newTemplate(name string) *template.Template {
	// create new template in the templates
	// set delimiters and add func map
	return r.templates.New(name).
		Funcs(r.FuncMap).
		Delims(r.Delims.Left, r.Delims.Right)
}

func (r *Renderer) loadTemplateFile(tplName, file string)  {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		panicErr(err)
	}

	// create new template in the templates
	template.Must(r.newTemplate(tplName).Parse(string(bs)))
}

/*************************************************************
 * render templates
 *************************************************************/

// Render a template name/file and write to the Writer.
// usage:
// 		renderer := view.NewRenderer()
//  	// ... ...
// 		// will apply global layout setting
// 		renderer.Render(http.ResponseWriter, "user/login", data)
// 		// apply custom layout file
// 		renderer.Render(http.ResponseWriter, "user/login", data, "custom-layout")
// 		// will disable apply layout render
// 		renderer.Render(http.ResponseWriter, "user/login", data, "")
func (r *Renderer) Render(w io.Writer, tplName string, data interface{}, layouts ...string) error {
	// apply layout
	if layout := r.getLayoutName(layouts); layout != "" {
		r.addLayoutFuncs(tplName, data)
		tplName = layout
	}

	return r.Partial(w, tplName, data)
}

// Partial render partial, will not render layout file
func (r *Renderer) Partial(w io.Writer, tplName string, data interface{}) error {
	tplName = r.CleanExt(tplName)

	// get a buffer from the pool to write to.
	buf := r.bufPool.get()
	err := r.templates.ExecuteTemplate(buf, tplName, data)
	if err != nil {
		return err
	}

	buf.WriteTo(w)
	r.bufPool.put(buf)

	return nil
}

// String render a template string
func (r *Renderer) String(w io.Writer, tplString string, data interface{}) error {
	t := template.New("").Delims(r.Delims.Left, r.Delims.Right).Funcs(r.FuncMap)
	template.Must(t.Parse(tplString))

	return t.Execute(w, data)
}

/*************************************************************
 * help methods
 *************************************************************/

func (r *Renderer) executeTemplate(name string, data interface{}) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	return buf, r.templates.ExecuteTemplate(buf, name, data)
}

// name the template name
func (r *Renderer) addLayoutFuncs(name string, data interface{}) {
	tpl := r.templates.Lookup(name)
	if tpl == nil {
		return
	}

	includeHandler := func(partialName string) (template.HTML, error) {
		if r.Template(partialName) != nil {
			buf, err := r.executeTemplate(partialName, data)
			// Return safe HTML here since we are rendering our own template.
			return template.HTML(buf.String()), err
		}

		return "", nil
	}

	funcMap := template.FuncMap{
		"yield": func() (template.HTML, error) {
			buf, err := r.executeTemplate(name, data)
			// Return safe HTML here since we are rendering our own template.
			return template.HTML(buf.String()), err
		},
		// current template name
		"current": func() (string, error) {
			return name, nil
		},
		"include": includeHandler,
		"partial": includeHandler,
	}

	tpl.Funcs(funcMap)
}
