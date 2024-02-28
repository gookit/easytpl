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

	r := easytpl.NewExtends(easytpl.WithDebug, easytpl.DisableLayout)
	r.LoadString("base", `{{ block "header" . }}header{{ end }}
{{ block "body" . }} default body{{ end }}
{{ block "footer" . }}footer{{ end }}`)
	r.LoadBytes("home", []byte(`{{ extends "base" }}
{{ define "body" }} body: hi, {{.}} {{ end }}`))

	t.Run("render base", func(t *testing.T) {
		err := r.Execute(bf, "base", "inhere")
		is.Nil(err)
		is.Equal("header\n default body\nfooter", bf.String())
		bf.Reset()
	})

	t.Run("render home", func(t *testing.T) {
		err := r.Execute(bf, "home", "inhere")
		is.Nil(err)
		is.Equal("header\n body: hi, inhere \nfooter", bf.String())
	})
}

func Example_extends() {
	r := easytpl.NewExtends()
	// load root
	r.LoadStrings(map[string]string{
		// layout template file
		"layout.tpl": `{{ block "header" . }}header{{ end }}
{{ block "body" . }}default{{ end }}
{{ block "footer" . }}footer{{ end }}`,
		// current page template
		"home.tpl": `{{ extends "layout.tpl" }}
{{ define "body" }}hello {{.}}{{ end }}`,
	})

	fmt.Println("- render 'layout.tpl'")
	err := r.Execute(os.Stdout, "layout.tpl", "inhere")
	if err != nil {
		panic(err)
	}

	fmt.Println("\n- render 'home.tpl'")
	err = r.Execute(os.Stdout, "home.tpl", "inhere")
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
	// hello inhere
	// footer
}
