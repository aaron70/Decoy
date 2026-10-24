package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/aaron70/decoy/pkg/decoy"
)

func main() {
	d, err := decoy.NewDecoy()
	if err != nil {
		fmt.Printf("Couldn't create the Decoy instance\n")
		return
	}

	template := `{{- $id := nextIncrementalInt "id" 0 1 -}}
	{{- $name := randomName -}}
	{ "id": {{ $id }}, "name": "{{ $name }}", "greeting": "{{ template "greeting" . }}" }`

	greeting := `Hello my name is {{ .name }}`

	compiledTemplate, cleanup, err := d.CompileTemplate(template, decoy.WithExtraTemplate("greeting", greeting))
	if err != nil {
		fmt.Printf("Couldn't compile the template: %s\n", err)
		return
	}
	defer cleanup()

	var wg sync.WaitGroup
	defer wg.Wait()

	data := map[string]any{
		"name": "loco",
	}

	for i := range 10 {
		wg.Go(func() {
			str := strings.Builder{}
			if err := compiledTemplate.Execute(&str, &data); err != nil {
				fmt.Printf("Error while executing the template (%d): %s\n", i, err)
			} else {
				fmt.Println(str.String())
			}
		})
	}
}
