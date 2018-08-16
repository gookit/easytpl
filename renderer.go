package view

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
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

// Renderer definition
type Renderer struct {
	extMap map[string]uint8
	// loaded files. {"tpl name": "file path"}
	fileMap map[string]string
	bufPool *bufferPool
	// templates is root template instance.
	// it is like a map, contains all parsed templates
	// {
	// 	"tpl name0": *template.Template,
	// 	"tpl name1": *template.Template,
	// 	... ...
	// }
	templates *template.Template
	// mark renderer is initialized
	initialized bool

	// Debug setting
	Debug bool
	// ViewsDir the default views directory
	ViewsDir string
	// Layout template name
	Layout string
	// Delims define for template
	Delims TplDelims
	// ExtNames allowed template extensions. eg {"tpl", "html"}
	ExtNames []string
	// FuncMap func map for template
	FuncMap template.FuncMap
	// DisableLayout disable layout. default is False
	DisableLayout bool
	// AutoSearchFile TODO)auto search template file, when not found on compiled templates. default is False
	AutoSearchFile bool
}

// NewRenderer create a new view renderer
func NewRenderer(fns ...func(r *Renderer)) *Renderer {
	r := &Renderer{
		Delims:   TplDelims{"{{", "}}"},
		FuncMap:  make(template.FuncMap),
		ExtNames: []string{"tpl", "html"},
		bufPool:  newBufferPool(),
		fileMap:  make(map[string]string),
	}

	// apply config func
	if len(fns) == 1 {
		fns[0](r)
	}

	return r
}

// NewInitialized create a new and initialized view renderer
func NewInitialized(fns ...func(r *Renderer)) *Renderer {
	r := NewRenderer(fns...)
	r.MustInitialize()
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

	if r.initialized {
		panicErr(fmt.Errorf("cannot add template func after initialized"))
	}

	if r.FuncMap == nil {
		r.FuncMap = make(template.FuncMap)
	}

	r.FuncMap[name] = fn
}

// AddFuncMap add template func map
func (r *Renderer) AddFuncMap(fm template.FuncMap) {
	if r.FuncMap == nil {
		r.FuncMap = make(template.FuncMap)
	}

	for name, fn := range fm {
		r.FuncMap[name] = fn
	}
}

// Initialize templates in the viewsDir, add do some prepare works.
// Notice: must call it on after create Renderer
func (r *Renderer) Initialize() error {
	if r.initialized {
		return nil
	}

	r.debugf("begin initialize the view renderer")
	if len(r.ExtNames) == 0 {
		r.ExtNames = []string{DefaultExt}
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

			str, err := r.executeTemplate(tplName, v)
			return template.HTML(str), err
		}

		return "", nil
	})

	// compile templates
	if err := r.compileTemplates(); err != nil {
		return err
	}

	r.debugf("renderer initialize is complete, added template func: %d", len(r.FuncMap))
	r.initialized = true
	return nil
}

// MustInitialize compile templates and report error
func (r *Renderer) MustInitialize() {
	panicErr(r.Initialize())
}

/*************************************************************
 * load template files, strings
 *************************************************************/

// LoadByGlob load templates by glob
// usage:
// 		r.LoadByGlob("views/*")
// 		r.LoadByGlob("views/**/*")
func (r *Renderer) LoadByGlob(pattern string, baseDirs ...string) {
	r.ensureTemplates()
	paths, err := filepath.Glob(pattern)
	panicErr(err)

	if r.fileMap == nil {
		r.fileMap = make(map[string]string)
	}

	var baseDir, relPath string
	if len(baseDirs) == 1 {
		baseDir = baseDirs[0]
	}

	for _, path := range paths {
		ext := filepath.Ext(path)
		if !r.IsValidExt(ext) {
			continue
		}

		relPath = path
		if baseDir != "" {
			relPath, err = filepath.Rel(baseDir, path)
			panicErr(err)
		}

		name := filepath.ToSlash(relPath[0 : len(relPath)-len(ext)])
		r.loadTemplateFile(name, path)
	}
}

// LoadFiles load template files.
// usage:
// 		r.LoadFiles("path/file1.tpl", "path/file2.tpl")
func (r *Renderer) LoadFiles(files ...string) {
	r.ensureTemplates()

	if r.fileMap == nil {
		r.fileMap = make(map[string]string)
	}

	for _, file := range files {
		ext := filepath.Ext(file)
		if r.IsValidExt(ext) {
			name := filepath.ToSlash(file[0 : len(file)-len(ext)])
			r.loadTemplateFile(name, file)
		}
	}
}

// LoadString load named template string.
// usage:
//
func (r *Renderer) LoadString(tplName string, tplString string) {
	r.ensureTemplates()
	r.debugf("load named template string, name is: %s", tplName)

	// create new template in the templates
	template.Must(r.newTemplate(tplName).Parse(tplString))
}

// LoadStrings load multi named template strings
func (r *Renderer) LoadStrings(sMap map[string]string) {
	for name, tplStr := range sMap {
		r.LoadString(name, tplStr)
	}
}

func (r *Renderer) compileTemplates() error {
	// create root template engine
	r.ensureTemplates()

	dir := r.ViewsDir
	if dir == "" {
		return nil
	}

	r.debugf("will compile templates from the views dir: %s", dir)

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

	r.templates = template.New("ROOT")
}

func (r *Renderer) newTemplate(name string) *template.Template {
	// create new template in the templates
	// set delimiters and add func map
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
	r.debugf("load template file: %s, and template name is: %s", file, tplName)

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
	// apply layout render
	if layout := r.getLayoutName(layouts); layout != "" {
		r.addLayoutFuncs(layout, tplName, data)
		tplName = layout
	}

	return r.Partial(w, tplName, data)
}

// Partial render partial, will not render layout file
func (r *Renderer) Partial(w io.Writer, tplName string, data interface{}) error {
	if !r.initialized {
		panicErr(fmt.Errorf("please call Initialize(), before render template"))
	}

	str, err := r.executeTemplate(tplName, data)
	if err != nil {
		return err
	}

	w.Write([]byte(str))
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

// execute data render by template name
func (r *Renderer) executeTemplate(name string, data interface{}) (string, error) {
	// get a buffer from the pool to write to.
	buf := r.bufPool.get()
	name = r.cleanExt(name)

	// find template instance by name
	tpl := r.templates.Lookup(name)
	if tpl == nil {
		return "", fmt.Errorf("view renderer: the template [%s] is not found", name)
	}

	tpl.Funcs(template.FuncMap{
		// current template name
		"current": func() string {
			return name
		},
	})

	r.debugf("render the template: %s, override set func: current", name)
	err := tpl.Execute(buf, data)
	str := buf.String()

	// return buffer to pool
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
			str, err := r.executeTemplate(name, data)
			return template.HTML(str), err
		},
		// will add data to included template
		// "include": includeHandler,
		// "partial": includeHandler,
	}

	tpl.Funcs(funcMap)
}
