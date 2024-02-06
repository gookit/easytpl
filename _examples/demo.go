package main

import (
	"fmt"
	"net/http"

	"github.com/gookit/easytpl"
)

var v *easytpl.Renderer

// go run ./_examples/demo.go
func main() {
	// equals to call: view.NewRenderer() + r.MustInit()
	v = easytpl.NewInitialized(func(r *easytpl.Renderer) {
		// setting default layout
		r.Layout = "layout" // equals to "layout.tpl"
		// templates dir. will auto load on init.
		r.ViewsDir = "testdata"
	})

	// fmt.Println(v.TemplateNames(true)) // output a loaded template names

	// load named template by string
	// now, you can use "my-page" as an template name
	v.LoadString("my-page", "<h1>welcome {{.}}</h1>")

	// more ways for load templates
	// v.LoadByGlob("some/path/*", "some/path")
	// v.LoadFiles("path/file1.tpl", "path/file2.tpl")

	addRoutes()
	fmt.Println("Listening on http://127.0.0.1:9100")
	http.ListenAndServe(":9100", nil)
}

func addRoutes() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<h1>hello, welcome</h1>"))
	})

	http.HandleFunc("/layout", func(w http.ResponseWriter, r *http.Request) {
		v.Render(w, "home", "tom")
	})

	http.HandleFunc("/no-layout", func(w http.ResponseWriter, r *http.Request) {
		v.Partial(w, "home", "tom")
	})

	http.HandleFunc("/my-page", func(w http.ResponseWriter, r *http.Request) {
		v.Partial(w, "my-page", "tom") // welcome tom
	})

	http.HandleFunc("/tpl-names", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(v.TemplateNames(true)))
	})
}
