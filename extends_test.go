package easytpl_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/gookit/easytpl"
	"github.com/gookit/goutil/testutil/assert"
)

func _TestUseExtends(t *testing.T) {
	bf := new(bytes.Buffer)
	is := assert.New(t)

	r := easytpl.NewInitialized(func(r *easytpl.Renderer) {
		r.Debug = true
		// r.ViewsDir = "testdata/extends"
	})
	r.LoadStrings(map[string]string{
		"home": `{{ extends "layout" . }}
{{ define "body" }} hello {{.}}{{ end }}`,
		// layout file
		"layout": `{{ block "header" . }}header{{ end }}
{{ block "body" . }}default{{ end }}
{{ block "footer" . }}footer{{ end }}`,
	})

	err := r.Execute(bf, "layout", "inhere")
	is.Nil(err)
	is.Equal("header\n hello inhere\nfooter", bf.String())
	bf.Reset()

	err = r.Execute(bf, "home", "inhere")
	is.Nil(err)
	is.Equal("header\n hello \nfooter\n", bf.String())
}

func _Example_Extends() {
	r := easytpl.NewInitialized()
	// load root
	r.LoadStrings(map[string]string{
		// layout template file
		"layout.tpl": `{{ block "header" . }}header{{ end }}
{{ block "body" . }}default{{ end }}
{{ block "footer" . }}footer{{ end }}`,
		// current page template
		"home.tpl": `{{ extends "layout.tpl" . }}
{{ define "body" }}hello {{.}}{{ end }}`,
	})

	fmt.Println("- render 'layout.tpl'")
	err := r.Execute(os.Stdout, "layout", "inhere")
	if err != nil {
		panic(err)
	}

	fmt.Println("\n- render 'home.tpl'")
	err = r.Execute(os.Stdout, "home", "inhere")
	if err != nil {
		panic(err)
	}

	// Output:
	// - render 'layout.tpl'
	// header
	// default
	// footer
	// - render 'home.tpl'
	// header
	// default
	// footer
}
