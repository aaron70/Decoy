package main

import (
	"fmt"

	"github.com/aaron70/decoy/pkg/template"
	"github.com/aaron70/goaty/utils"
)

func main() {
	N := 5
	tmpl := utils.Must(template.NewTemplate(`Hello, {{ randomName }}, {{ nextIncrementalInt "id" 0 1 }}`))
	tmpl2 := utils.Must(template.NewTemplate(`Hello, {{ randomName }}, {{ nextIncrementalInt "id" 0 1 }}`))

	for range N {
		fmt.Printf("[1]: %s\n", utils.Must(tmpl.ParseString()))
		fmt.Printf("[2]: %s\n", utils.Must(tmpl2.ParseString()))
	}
}
