package easytpl

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/gookit/goutil/basefn"
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
	// from ViewsDir, split by comma
	tplDirs []string

	// Debug setting
	Debug bool
	// Layout template name
	Layout string
	// Delims define for template
	Delims TplDelims
	// ViewsDir the default views directory, multi dirs use "," split
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

// ConfigFn for renderer
type ConfigFn func(r *Renderer)

// NewRenderer create a new view renderer
func NewRenderer(fns ...ConfigFn) *Renderer {
	r := &Renderer{
		Delims:   TplDelims{"{{", "}}"},
		FuncMap:  make(template.FuncMap),
		ExtNames: []string{"tpl", "html"},
		bufPool:  newBufferPool(),
		fileMap:  make(map[string]string),
	}

	// Apply config func
	return r.WithConfig(fns...)
}

// NewInitialized create a new and initialized view renderer.
func NewInitialized(fns ...ConfigFn) *Renderer {
	r := NewRenderer(fns...)
	return r.MustInit()
}

// WithConfig apply config func
func (r *Renderer) WithConfig(fns ...ConfigFn) *Renderer {
	for _, fn := range fns {
		fn(r)
	}
	return r
}

/*************************************************************
 * prepare for view Renderer
 *************************************************************/

// AddFunc add template func
func (r *Renderer) AddFunc(name string, fn any) {
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
func (r *Renderer) AddFuncMap(fm template.FuncMap) {
	if r.FuncMap == nil {
		r.FuncMap = make(template.FuncMap)
	}

	for name, fn := range fm {
		r.FuncMap[name] = fn
	}
}

// MustInitialize compile templates and report error
//
// Deprecated: please use MustInit, will remove this method in the future
func (r *Renderer) MustInitialize() { r.MustInit() }

// MustInit compile templates, will panic on error
func (r *Renderer) MustInit() *Renderer {
	basefn.MustOK(r.Initialize())
	return r
}

// Initialize templates in the viewsDir, add do some prepare works.
//
// Notice: must call it on after create Renderer
func (r *Renderer) Initialize() error {
	if r.initialized {
		return nil
	}

	r.debugf("begin initialize the view renderer")
	if len(r.ExtNames) == 0 {
		r.ExtNames = []string{DefaultExt}
	}

	r.tplDirs = strings.Split(r.ViewsDir, ",")
	r.extMap = make(map[string]uint8, len(r.ExtNames))
	for _, ext := range r.ExtNames {
		if ext[0] != '.' {
			ext = "." + ext
		}

		r.extMap[ext] = 0
	}

	// add template func
	r.AddFuncMap(globalFuncMap)
	r.AddFunc("include", func(tplName string, data ...any) (template.HTML, error) {
		if tpl := r.Template(tplName); tpl != nil {
			var v any
			if len(data) == 1 {
				v = data[0]
			}

			str, err := r.executeTemplate(tpl, v)
			return template.HTML(str), err
		}

		return "", nil
	})
	r.AddFunc("extends", func(tplName string, data ...any) (template.HTML, error) {
		if tpl := r.Template(tplName); tpl != nil {
			var v any
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

/*************************************************************
 * load template files, strings
 *************************************************************/

// LoadByGlob load templates by glob pattern.
//
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
//
// Usage:
//
//	r.LoadFiles("path/file1.tpl", "path/file2.tpl")
func (r *Renderer) LoadFiles(files ...string) {
	if !r.initialized {
		panicErr(errors.New("please call Initialize(), before load templates"))
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
func (r *Renderer) LoadStrings(sMap map[string]string) {
	for name, tplStr := range sMap {
		r.LoadString(name, tplStr)
	}
}

/*************************************************************
 * internal helper methods
 *************************************************************/

func (r *Renderer) compileTemplates() error {
	// create root template engine
	r.ensureTemplates()

	for _, tplDir := range r.tplDirs {
		if err := r.compileInDir(tplDir); err != nil {
			return err
		}
	}
	return nil
}

func (r *Renderer) compileInDir(dir string) error {
	r.debugf("will compile templates from the views dir: %s", dir)

	// Walk the supplied directory and compile any files that match our extension list.
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// Skip dir
		if info == nil || info.IsDir() {
			return nil
		}

		// Path is full path, rel is relative path
		// e.g. path: "testdata/admin/footer.tpl" -> rel: "admin/footer.tpl"
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
	// will inherit delimiters and add func map
	return r.templates.New(name)
}

func (r *Renderer) loadTemplateFile(tplName, file string) {
	bs, err := os.ReadFile(file)
	if err != nil {
		panicErr(err)
	}

	r.fileMap[tplName] = file
	r.debugf("load template file: %s, template name is: %s", file, tplName)

	// Create new template in the templates
	template.Must(r.newTemplate(tplName).Parse(string(bs)))
}

// name the template name
func (r *Renderer) addLayoutFuncs(layout, name string, data any) {
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

/*************************************************************
 * Helper methods
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
