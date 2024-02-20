package easytpl

import (
	"bytes"
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
	Options
	bufPool *bufferPool
	// mark renderer is initialized
	init bool
	// root It is the root template instance.
	//
	// It is like a map, contains all parsed templates.
	//
	// {
	// 	"tpl name0": *template.Template,
	// 	"tpl name1": *template.Template,
	// 	... ...
	// }
	root *template.Template

	// from ViewsDir, split by comma
	tplDirs []string
	// supported template file extension names
	extMap map[string]uint8
	// loaded template files. {"tpl name": "file path"}
	fileMap map[string]string

	// parsed from tpl file first line.
	//
	// eg: home.tpl: {{ extends "some/base.tpl" }}
	// -> baseTpl = {"home": "some/base.tpl", }
	//
	// Note: this is for extends feature on Options.EnableExtends is True
	baseTpl map[string]string
}

// NewRenderer create a new view renderer
func NewRenderer(fns ...OptionFn) *Renderer {
	r := &Renderer{
		bufPool: newBufferPool(),
		fileMap: make(map[string]string),
		Options: Options{
			Delims:   TplDelims{"{{", "}}"},
			FuncMap:  make(template.FuncMap),
			ExtNames: []string{"tpl", "html"},
		},
	}

	return r.WithOptions(fns...)
}

// WithOptions apply config functions
func (r *Renderer) WithOptions(fns ...OptionFn) *Renderer {
	for _, fn := range fns {
		fn(r)
	}
	return r
}

/*************************************************************
 * prepare for template Renderer
 *************************************************************/

// AddFunc add template func
func (r *Renderer) AddFunc(name string, fn any) {
	r.cannotInit("cannot add template func after initialized")
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		panicf("the template func [%s] type must be a function", name)
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
func (r *Renderer) Initialize() error { return r.Init() }

// Init Initialize templates in the viewsDir, add do some prepare works.
//
// Notice: must call it on after create Renderer
func (r *Renderer) Init() error {
	if r.init {
		return nil
	}

	r.debugf("begin initialize the renderer: init fields, add funcs ...")
	if len(r.ExtNames) == 0 {
		r.ExtNames = []string{DefaultExt}
	}
	if len(r.ViewsDir) > 0 {
		r.tplDirs = strings.Split(r.ViewsDir, ",")
	}

	r.extMap = make(map[string]uint8, len(r.ExtNames))
	for _, ext := range r.ExtNames {
		if ext[0] != '.' {
			ext = "." + ext
		}
		r.extMap[ext] = 0
	}

	// add custom template func
	r.AddFuncMap(globalFuncMap)
	r.AddFunc("include", r.handleInclude)
	// r.AddFunc("extends", r.handleExtends)

	r.init = true
	// compile templates
	if err := r.compileTemplates(); err != nil {
		return err
	}

	r.debugf("renderer initialize is complete, added template func: %d", len(r.FuncMap))
	return nil
}

/*************************************************************
 * load template files, strings
 *************************************************************/

// LoadByGlob load templates by glob pattern. will panic on error
//
// Usage:
//
//	r.LoadByGlob("views/*")
//	r.LoadByGlob("views/*", "views") // register template will remove prefix "views"
//	r.LoadByGlob("views/*.tpl") // add ext limit
//	r.LoadByGlob("views/**/*") // all sub-dir files
func (r *Renderer) LoadByGlob(pattern string, baseDirs ...string) {
	r.requireInit("must call Initialize() before load templates")

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

		// name: path without extension
		name := filepath.ToSlash(relPath[0 : len(relPath)-len(ext)])
		r.LoadFile(name, path)
	}
}

// LoadFiles load custom template files.
//
// Usage:
//
//	r.LoadFiles("path/file1.tpl", "path/file2.tpl")
func (r *Renderer) LoadFiles(files ...string) {
	r.requireInit("must call Initialize() before load templates")

	for _, file := range files {
		ext := filepath.Ext(file)
		if r.IsValidExt(ext) {
			// name: path without extension
			name := filepath.ToSlash(file[0 : len(file)-len(ext)])
			r.LoadFile(name, file)
		}
	}
}

// LoadFile load named template file. will panic on error
func (r *Renderer) LoadFile(tplName, filePath string) {
	bs, err := os.ReadFile(filePath)
	panicErr(err)

	r.fileMap[tplName] = filePath
	r.debugf("load template file: %s, template name: %s", filePath, tplName)
	r.loadBytes(tplName, bs)
}

// LoadString load named template string. will panic on error
//
// Usage:
//
//	r.LoadString("my-page", "welcome {{.}}")
//	// now, you can use "my-page" as a template name
//	r.Partial(w, "my-page", "tom") // Result: "welcome tom"
func (r *Renderer) LoadString(tplName, tplText string) {
	r.debugf("load named template text, name is: %s", tplName)
	r.loadBytes(tplName, []byte(tplText))
}

// LoadStrings load multi named template strings.
// key is template name, value is template contents.
func (r *Renderer) LoadStrings(sMap map[string]string) {
	for name, tplStr := range sMap {
		r.LoadString(name, tplStr)
	}
}

// LoadBytes load named template bytes. will panic on error
func (r *Renderer) LoadBytes(tplName string, tplText []byte) {
	r.debugf("load named template bytes, name is: %s", tplName)
	r.loadBytes(tplName, tplText)
}

func (r *Renderer) loadBytes(tplName string, tplText []byte) {
	r.ensureRoot()

	// parse the first line of the text, collect the base template name
	if r.EnableExtends {
		bs := tplText
		bs = bytes.TrimLeft(bs, "\n\t ")

		if i := bytes.IndexByte(bs, '\n'); i >= 0 {
			if baseTpl, ok := getExtendsTplName(bs[0:i], r.Delims); ok {
				r.baseTpl[tplName] = baseTpl
				bs = bs[i+1:] // remove the first line
			}
		}
	}

	// create new template in the root, will inherit delimiters and all func map
	template.Must(r.root.New(tplName).Parse(string(tplText)))
}

/*************************************************************
 * internal helper methods
 *************************************************************/

func (r *Renderer) ensureRoot() {
	r.requireInit("must call Initialize() before current operation")
	if r.root == nil {
		r.root = template.New("ROOT").Funcs(r.FuncMap)
	}

	if r.fileMap == nil {
		r.fileMap = make(map[string]string)
	}
}

func (r *Renderer) compileTemplates() error {
	r.ensureRoot()

	for _, tplDir := range r.tplDirs {
		if err := r.compileInDir(tplDir); err != nil {
			return err
		}
	}
	return nil
}

func (r *Renderer) compileInDir(dir string) error {
	r.debugf("will compile templates in the dir: %s", dir)

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

		// skip no extension file
		ext := filepath.Ext(rel)
		if len(ext) == 0 {
			return nil
		}

		// load on is supported extension
		if _, has := r.extMap[ext]; has {
			name := rel[0 : len(rel)-len(ext)]
			r.LoadFile(name, path)
		}
		return err
	})
}

// name the template name
func (r *Renderer) addLayoutFuncs(layout, name string, data any) {
	tpl := r.Template(layout)
	if tpl == nil {
		panicf("the layout template %q is not found, want render: %s", layout, name)
	}

	// includeHandler := func(tplName string) (template.HTML, error) {
	// 	if r.root.Lookup(tplName) != nil {
	// 		str, err := r.executeTemplate(tplName, data)
	// 		// Return safe HTML here since we are rendering our own template.
	// 		return template.HTML(str), err
	// 	}
	//
	// 	return "", nil
	// 	}

	r.debugf("add funcs[yield, partial] to layout template: %s, target template: %s", layout, name)
	tpl.Funcs(template.FuncMap{
		"yield": func() (template.HTML, error) {
			str, err := r.executeByName(name, data)
			return template.HTML(str), err
		},
		// Will add data to included template
		// "include": includeHandler,
		// "partial": includeHandler,
	})
}

/*************************************************************
 * Helper methods
 *************************************************************/

// Templates returns loaded template instances, including ROOT itself.
func (r *Renderer) Templates() []*template.Template {
	return r.root.Templates()
}

// TemplateFiles returns loaded template files
func (r *Renderer) TemplateFiles() map[string]string {
	return r.fileMap
}

var nameRpl = strings.NewReplacer(":", ":\n", ",", "\n")

// TemplateNames returns loaded template names
func (r *Renderer) TemplateNames(formatted ...bool) string {
	str := r.root.DefinedTemplates()
	if len(formatted) != 1 || formatted[0] == false {
		return str
	}
	return nameRpl.Replace(strings.TrimLeft(str, "; "))
}

// Root returns root template instance
func (r *Renderer) Root() *template.Template { return r.root }

// Template get template instance by name
func (r *Renderer) Template(name string) *template.Template {
	return r.root.Lookup(r.cleanExt(name))
}

// IsValidExt check is valid ext name
func (r *Renderer) IsValidExt(ext string) bool {
	_, ok := r.extMap[ext]
	return ok
}

// CleanExt will clean file ext.
//
// eg:
//
//	"some.tpl" -> "some"
//	"path/some.tpl" -> "path/some"
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

func (r *Renderer) getLayoutName(tplNames []string) string {
	var layout string
	disableLayout := r.DisableLayout

	if len(tplNames) > 0 {
		layout = strings.TrimSpace(tplNames[0])
		if layout == "" {
			disableLayout = true
		}
	} else {
		layout = r.Layout
	}

	if disableLayout {
		return ""
	}
	return layout
}

func (r *Renderer) requireInit(format string, args ...any) {
	if !r.init {
		panicf(format, args...)
	}
}

func (r *Renderer) cannotInit(format string, args ...any) {
	if r.init {
		panicf(format, args...)
	}
}

func (r *Renderer) debugf(format string, args ...any) {
	if r.Debug {
		fmt.Printf("easytpl: [DEBUG] "+format+"\n", args...)
	}
}
