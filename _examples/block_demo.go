package main

import (
	"fmt"
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

var (
	funcs     = template.FuncMap{"join": strings.Join}
	guardians = []string{"Gamora", "Groot", "Nebula", "Rocket", "Star-Lord"}
)

var root = template.New("root").Funcs(funcs)

func main() {
	fromGoDoc()
	// blockDemo()
}

func blockDemo() {
	// add master
	template.Must(root.New("master").Parse(master))
	masterTmpl := root.Lookup("master")

	// NOTE: 必须clone一个新的base template，否则会报错
	subTmpl, err := template.Must(masterTmpl.Clone()).Parse(overlay)
	goutil.PanicErr(err)

	overlayTmpl, err := root.AddParseTree("overlay", subTmpl.Tree)
	goutil.PanicErr(err)

	fmt.Println(
		root.DefinedTemplates(), "\n",
		subTmpl.DefinedTemplates(), "\n",
		overlayTmpl.DefinedTemplates(), "\n",
		root.Lookup("overlay").DefinedTemplates(),
	)

	// overlayTmpl := root.Lookup("overlay")

	fmt.Println("-------- render master --------")
	goutil.PanicErr(masterTmpl.Execute(os.Stdout, guardians))

	fmt.Println("-------- render overlay --------")
	goutil.PanicErr(overlayTmpl.Execute(os.Stdout, guardians))
	fmt.Println("")
}

func fromGoDoc() {
	// masterTmpl, err := template.New("master").Funcs(funcs).Parse(master)
	masterTmpl, err := root.New("master").Parse(master)
	goutil.PanicErr(err)

	// NOTE: 必须clone一个新的base template，否则会报错
	overlayTmpl, err := template.Must(masterTmpl.Clone()).Parse(overlay)
	overlayTmpl.Tree.Name = "overlay"
	overlayTmpl.Tree.ParseName = "overlay"
	goutil.PanicErr(err)

	fmt.Println("render master:", masterTmpl.Name(), masterTmpl.ParseName)
	goutil.PanicErr(masterTmpl.Execute(os.Stdout, guardians))

	fmt.Println("\nrender overlay:", overlayTmpl.Name(), overlayTmpl.ParseName, overlayTmpl.Tree.Name)
	goutil.PanicErr(overlayTmpl.Execute(os.Stdout, guardians))
}
