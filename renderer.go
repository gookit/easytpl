package easytpl

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"reflect"
	"strings"
)

// Renderer definition
type Renderer struct {
	extMap map[string]uint8
	// loaded files. {"tpl name": "file path"}
	fileMap map[string]string
	bufPool *bufferPool
	// templates are root template instance.
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
	// Layout template name
	Layout string
	// Delims define for template
	Delims TplDelims
	// ViewsDir the default views directory
	ViewsDir string
	// ExtNames allowed template extensions. eg {"tpl", "html"}
	ExtNames []string
	// FuncMap func map for template
	FuncMap template.FuncMap
	// DisableLayout disable layout. default is False
	DisableLayout bool
	// AutoSearchFile
	// TODO: auto search template file, when not found on compiled templates. default is False
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

	// Apply config func
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
 * prepare for view Renderer
 *************************************************************/

// AddFunc add template func
func AddFunc(name string, fn interface{}) {
	std.AddFunc(name, fn)
}

// AddFunc add template func
func (r *Renderer) AddFunc(name string, fn interface{}) {
	if r.initialized {
		panicErr(fmt.Errorf("cannot add template func after initialized"))
	}

	if reflect.TypeOf(fn).Kind() != reflect.Func {
		panicErr(fmt.Errorf("the template [%s] func must be a callable function", name))
	}

	if r.FuncMap == nil {
		r.FuncMap = make(template.FuncMap)
	}

	r.FuncMap[name] = fn
}

// AddFuncMap add template func map
func AddFuncMap(fm template.FuncMap) {
	std.AddFuncMap(fm)
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

// Initialize the default instance
func Initialize(fns ...func(r *Renderer)) {
	// Apply config func
	if len(fns) == 1 {
		fns[0](std)
	}

	std.MustInitialize()
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
		if tpl := r.Template(tplName); tpl != nil {
			var v interface{}
			if len(data) == 1 {
				v = data[0]
			}

			str, err := r.executeTemplate(tpl, v)
			return template.HTML(str), err
		}

		return "", nil
	})
	r.AddFunc("extends", func(tplName string, data ...interface{}) (template.HTML, error) {
		if tpl := r.Template(tplName); tpl != nil {
			var v interface{}
			if len(data) == 1 {
				v = data[0]
			}

			// NOTICE: must use a clone instance
			str, err := r.executeTemplate(tpl, v)
			return template.HTML(str), err
		}

		return "", nil
	})

	if r.fileMap == nil {
		r.fileMap = make(map[string]string)
	}

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

// LoadByGlob load templates by glob pattern.
func LoadByGlob(pattern string, baseDirs ...string) {
	std.LoadByGlob(pattern, baseDirs...)
}

// LoadByGlob load templates by glob pattern.
// Usage:
//
//	r.LoadByGlob("views/*")
//	r.LoadByGlob("views/*.tpl") // add ext limit
//	r.LoadByGlob("views/**/*")
func (r *Renderer) LoadByGlob(pattern string, baseDirs ...string) {
	if !r.initialized {
		panicErr(fmt.Errorf("please call Initialize(), before load templates"))
	}

	paths, err := filepath.Glob(pattern)
	panicErr(err)

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

// LoadFiles load custom template files.
func LoadFiles(files ...string) {
	std.LoadFiles(files...)
}

// LoadFiles load custom template files.
// Usage:
//
//	r.LoadFiles("path/file1.tpl", "path/file2.tpl")
func (r *Renderer) LoadFiles(files ...string) {
	if !r.initialized {
		panicErr(fmt.Errorf("please call Initialize(), before load templates"))
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
func LoadString(tplName string, tplString string) {
	std.LoadString(tplName, tplString)
}

// LoadString load named template string.
// Usage:
//
//	r.LoadString("my-page", "welcome {{.}}")
//	// now, you can use "my-page" as an template name
//	r.Partial(w, "my-page", "tom") // welcome tom
func (r *Renderer) LoadString(tplName string, tplString string) {
	r.ensureTemplates()
	r.debugf("load named template string, name is: %s", tplName)

	// Create new template in the templates
	template.Must(r.newTemplate(r.cleanExt(tplName)).Parse(tplString))
}

// LoadStrings load multi named template strings
func LoadStrings(sMap map[string]string) {
	std.LoadStrings(sMap)
}

// LoadStrings load multi named template strings
func (r *Renderer) LoadStrings(sMap map[string]string) {
	for name, tplStr := range sMap {
		r.LoadString(name, tplStr)
	}
}

/*************************************************************
 * render templates
 *************************************************************/

// Render a template name/file and write to the Writer.
func Render(w io.Writer, tplName string, v interface{}, layouts ...string) error {
	return std.Render(w, tplName, v, layouts...)
}

// Render a template name/file and write to the Writer.
//
// Usage:
//
//			renderer := easytpl.NewRenderer()
//	 	// ... ...
//			// will apply global layout setting
//			renderer.Render(http.ResponseWriter, "user/login", data)
//			// apply custom layout file
//			renderer.Render(http.ResponseWriter, "user/login", data, "custom-layout")
//			// will disable apply layout render
//			renderer.Render(http.ResponseWriter, "user/login", data, "")
func (r *Renderer) Render(w io.Writer, tplName string, v interface{}, layouts ...string) error {
	// Apply layout render
	if layout := r.getLayoutName(layouts); layout != "" {
		r.addLayoutFuncs(layout, tplName, v)
		tplName = layout
	}

	return r.Execute(w, tplName, v)
}

// Partial is alias of the Execute()
func Partial(w io.Writer, tplName string, v interface{}) error {
	return std.Execute(w, tplName, v)
}

// Partial is alias of the Execute()
func (r *Renderer) Partial(w io.Writer, tplName string, v interface{}) error {
	return r.Execute(w, tplName, v)
}

// Execute render partial, will not render layout file
func Execute(w io.Writer, tplName string, v interface{}) error {
	return std.Execute(w, tplName, v)
}

// Execute render partial, will not render layout file
func (r *Renderer) Execute(w io.Writer, tplName string, v interface{}) error {
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
func String(w io.Writer, tplString string, v interface{}) error {
	return std.String(w, tplString, v)
}

// String render a template string
func (r *Renderer) String(w io.Writer, tplString string, v interface{}) error {
	t := template.New("").Delims(r.Delims.Left, r.Delims.Right).Funcs(r.FuncMap)

	template.Must(t.Parse(tplString))
	return t.Execute(w, v)
}

/*************************************************************
 * Getter methods
 *************************************************************/

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

// IsValidExt check is valid ext name
func (r *Renderer) IsValidExt(ext string) bool {
	_, ok := r.extMap[ext]
	return ok
}
