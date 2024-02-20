package main

import (
	"os"
	"strings"
	"text/template"

	"github.com/gookit/goutil"
)

const (
	master = `Names:{{block "list" .}}
{{range .}}{{println "-" .}}{{end}}{{end}}`
	overlay = `{{define "list"}} {{join . ", "}}{{end}} `
)

func main() {
	var (
		funcs     = template.FuncMap{"join": strings.Join}
		guardians = []string{"Gamora", "Groot", "Nebula", "Rocket", "Star-Lord"}
	)

	// masterTmpl, err := template.New("master").Funcs(funcs).Parse(master)
	root := template.New("root").Funcs(funcs)
	masterTmpl, err := root.New("master").Parse(master)
	goutil.PanicErr(err)

	overlayTmpl, err := template.Must(masterTmpl.Clone()).Parse(overlay)
	goutil.PanicErr(err)

	goutil.PanicErr(masterTmpl.Execute(os.Stdout, guardians))
	goutil.PanicErr(overlayTmpl.Execute(os.Stdout, guardians))
}
