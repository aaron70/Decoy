package runner

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aaron70/decoy/internal/services"
	"github.com/aaron70/decoy/internal/utils"
	"github.com/spf13/cobra"
)

func CreateRunCommand(decoy *services.Decoy) *cobra.Command {
	var (
		extraTemplatesNames         []string
		_type                    services.RunnerType
		tmpl, file, data, config string
		stringValues             map[string]string
		n, g                     int
		err                      error
	)
	command := &cobra.Command{
		Use:   "run <type> [<runner>] [<template>]",
		Short: "Executes the runner of the given type.",
		Long: `Executes the runner of the given type using a runner config and a template.
The runner type can be: http, cmd.

You can provide a runner config by name or inline via --config/-c.
You can provide a template by name, inline via --template/-t, from a file via --file/-f, or from stdin.`,
		Example: `# Run a saved runner and saved template
decoy runner run cmd "echo" "greet"
decoy run http "http-post" "user-template" -n 10 -g 3

# Run with a saved runner and inline template
decoy runner run http "http-poster" -t '{"name": "{{ coalesce .Name "Doe" }}", "age": {{ randomInt 18 99 }} }' -n 5
decoy run http "http-poster" -t '{"name": "{{ coalesce .Name "Doe" }}", "age": {{ randomInt 18 99 }} }' -n 5

# Run with an inline config and inline template
decoy runner run http -c 'echo {{.template}}' -t '{"id": {{randomInt 1 100}}}'
decoy run http -c '{"method":"GET","url":"http://localhost:8080/{{ nextIncrementalInt \"id\" 0 1 }}"}'

# Run with data and values
decoy runner run http "http-poster" "user-template" --data '{"env":"test"}' -v region=us
decoy run http "http-poster" "user-template" --data '{"env":"test"}' -v region=us
`,
		Args: cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			_type = services.RunnerType(args[0])

			if len(args) >= 2 {
				entity, err := decoy.RunnerSvc.Get(args[1])
				if err != nil {
					return err
				}
				config = entity.Config
			} else if !cmd.Flags().Changed("config") {
				fmt.Printf("reading from stdin\n")
				tmpl, err = utils.ReadStringFrom(cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("Couldn't read the runner's config from stdin: %w", err)
				}
			}

			if len(args) == 3 {
				entity, err := decoy.TemplateSvc.Get(args[2])
				if err != nil {
					return err
				}
				tmpl = entity.Tmpl
			} else if cmd.Flags().Changed("file") {
				bytes, err := os.ReadFile(file)
				if err != nil {
					return err
				}
				tmpl = string(bytes)
			}

			var jsonData map[string]any
			err = json.Unmarshal([]byte(data), &jsonData)
			if err != nil {
				return fmt.Errorf("Couldn't deserialize the contents from the flag --data/-d: %w", err)
			}

			for key, value := range stringValues {
				jsonData[key] = value
			}

			extraTemplates := make(map[string]string, len(extraTemplatesNames))
			for _, extraTemplate := range extraTemplatesNames {
				extraTmpl, err := decoy.TemplateSvc.Get(extraTemplate)
				if err != nil {
					return err
				}
				extraTemplates[extraTemplate] = extraTmpl.Tmpl
			}

			return decoy.RunnerSvc.Run(cmd.OutOrStdout(), _type, config, tmpl, extraTemplates, jsonData, n, g)
		},
	}

	command.Flags().StringVarP(&tmpl, "template", "t", "", "The content of the template to parse")
	command.Flags().StringVarP(&file, "file", "f", "", "The path of the file with the template contents to parse")
	command.MarkFlagsMutuallyExclusive("template", "file")

	command.Flags().StringSliceVarP(&extraTemplatesNames, "extraTemplate", "x", make([]string, 0, 5), "Extra templates to parse so they can be used (nested) within the main template")

	command.Flags().StringVarP(&data, "data", "d", "{}", "The JSON data to be used within the template")
	command.Flags().StringToStringVarP(&stringValues, "value", "v", map[string]string{}, "A set of pairs (key=value) to inject into the data")

	command.Flags().IntVarP(&n, "times", "n", 1, "The number of times the runner will be executed")
	command.Flags().IntVarP(&g, "goroutines", "g", 1, "The number of concurrent goroutines executing the runner")

	command.Flags().StringVarP(&config, "config", "c", "", "The config's content of the runner to execute")

	return command
}
