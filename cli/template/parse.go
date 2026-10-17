package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aaron70/decoy/internal/services"
	"github.com/aaron70/decoy/internal/utils"
	"github.com/aaron70/decoy/pkg/decoy"
	"github.com/spf13/cobra"
)

func CreateParseCommand(cli *services.Decoy) *cobra.Command {
	var (
		extraTemplatesNames    []string
		name, tmpl, file, data string
		err                    error
		stringValues           map[string]string
	)

	command := &cobra.Command{
		Use:   "parse [<name>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Parses the given template and prints the results to stdout.",
		Example: `# Parse a saved template
decoy template parse "greet" --data '{ "Name": "Doe" }'
decoy parse "greet" -v Name=Doe

# Parse an inline template
decoy template parse -t 'Hello, {{ coalesce .Name "World" }}!' --data '{ "Name": "Doe" }'
decoy parse -t 'Hello, {{ coalesce .Name "World" }}!' -v Name=Doe

# Parse a template from stdin
echo 'Hello, {{ coalesce .Name "World" }}!' | decoy template parse --data '{ "Name": "Doe" }'
echo 'Hello, {{ coalesce .Name "World" }}!' | decoy parse -v Name=Doe

# Parse a template from a file
decoy template parse -f /path/to/template --data '{ "Name": "Doe" }'
decoy parse -f /path/to/template -v Name=Doe
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				name = args[0]
				if tmpl != "" {
					fmt.Fprintf(cmd.ErrOrStderr(), "[Warning]: Ignoring --template/-t flag since the template name %q was provided.", name)
				}
				if file != "" {
					fmt.Fprintf(cmd.ErrOrStderr(), "[Warning]: Ignoring --file/-f flag since the template name %q was provided.", name)
				}
				entity, err := cli.TemplateSvc.Get(name)
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
			} else if !cmd.Flags().Changed("template") {
				tmpl, err = utils.ReadStringFrom(cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("Couldn't read the template from stdin: %w", err)
				}
			}

			var jsonData map[string]any
			err := json.Unmarshal([]byte(data), &jsonData)
			if err != nil {
				return fmt.Errorf("Couldn't deserialize the contents from the flag --data/-d: %w", err)
			}

			for key, value := range stringValues {
				jsonData[key] = value
			}

			parsedTemplate := bytes.NewBufferString("")

			extraTemplates := make(map[string]string, len(extraTemplatesNames))
			for _, extraTemplate := range extraTemplatesNames {
				extraTmpl, err := cli.TemplateSvc.Get(extraTemplate)
				if err != nil {
					return err
				}
				extraTemplates[extraTemplate] = extraTmpl.Tmpl
			}

			err = cli.Decoy.ParseTemplate(parsedTemplate, tmpl,
				decoy.WithFuncMap(cli.Decoy.DefaultTemplateFuncMaps()),
				decoy.WithData(jsonData),
				decoy.WithName(name),
				decoy.WithExtraTemplates(extraTemplates),
			)
			if !strings.HasSuffix(parsedTemplate.String(), "\n") {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", parsedTemplate.String())
			} else {
				fmt.Fprint(cmd.OutOrStdout(), parsedTemplate.String())
			}
			return err
		},
	}

	command.Flags().StringVarP(&tmpl, "template", "t", "", "The content of the template to parse")
	command.Flags().StringVarP(&file, "file", "f", "", "The path of the file with the template contents to parse")
	command.MarkFlagsMutuallyExclusive("template", "file")

	command.Flags().StringSliceVarP(&extraTemplatesNames, "extraTemplate", "x", make([]string, 0, 5), "Extra templates to parse so they can be used (nested) within the main template")

	command.Flags().StringVarP(&data, "data", "d", "{}", "The JSON data to be used within the template")
	command.Flags().StringToStringVarP(&stringValues, "value", "v", map[string]string{}, "A set of pairs (key=value) to inject into the data")

	return command
}
