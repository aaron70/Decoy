package main

import (
	"fmt"
	"log"

	"github.com/aaron70/decoy/pkg/template"
)

func main() {
	tmpl, err := template.NewTemplate(
		`Hello, {{ .Name }}, {{ randomName }}, {{ nextIncrementalInt "id" 0 1 }}`,
		template.WithDefaultNgrams(false),
	)
	if err != nil {
		log.Fatal(err)
	}

	for range 5 {
		text, err := tmpl.ParseString(template.WithData(map[string]string{"Name": "Aaron"}))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(text)
	}
}
