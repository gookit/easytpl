package easytpl_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/gookit/easytpl"
	"github.com/gookit/goutil/testutil/assert"
)

func TestUseExtends(t *testing.T) {
	bf := new(bytes.Buffer)
	is := assert.New(t)

	r := easytpl.NewInited(easytpl.WithDebug, func(r *easytpl.Renderer) {
		r.DisableLayout = true
	})
	r.LoadStrings(map[string]string{
		"home": `{{ extends "layout" . }}
{{ define "body" }} body: hi, {{.}}{{ end }}`,
		// layout file
		"layout": `{{ block "header" . }}header{{ end }}
{{ block "body" . }} default body{{ end }}
{{ block "footer" . }}footer{{ end }}`,
	})

	// t.Run("render layout", func(t *testing.T) {
	// 	err := r.Execute(bf, "layout", "inhere")
	// 	is.Nil(err)
	// 	is.Equal("header\n default body\nfooter", bf.String())
	// 	bf.Reset()
	// })

	t.Run("render home", func(t *testing.T) {
		err := r.Execute(bf, "home", "inhere")
		fmt.Println(err)
		is.Nil(err)
		is.Equal("header\n body: hi, inhere \nfooter\n", bf.String())
	})
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
