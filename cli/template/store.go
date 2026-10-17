package template

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aaron70/decoy/internal/services"
	"github.com/aaron70/decoy/internal/utils"
	errs "github.com/aaron70/goaty/errors"
	"github.com/spf13/cobra"
)

func createStoreCommand(decoy *services.Decoy) *cobra.Command {
	var (
		name, tmpl, file string
		err              error
	)
	command := &cobra.Command{
		Use:   "store <name>",
		Short: "Upserts the contents of a template.",
		Long:  "Upserts the contents of a template. You can pass the template's content from stdin, a file, or through the template flag.",
		Example: `# Store a template from stdin
echo 'Hello, {{ coalesce .Name "World" }}' | decoy template store "greet"

# Store a template from a file
decoy template store "greet" -f /path/to/template

# Store an inline template
decoy template store "greet" -t 'Hello, {{ coalesce .Name "World" }}'
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name = args[0]
			if cmd.Flags().Changed("file") {
				bytes, err := os.ReadFile(file)
				if err != nil {
					return err
				}
				tmpl = strings.TrimRight(string(bytes), "\n")
			} else if !cmd.Flags().Changed("template") {
				tmpl, err = utils.ReadStringFrom(cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("Couldn't read the template from stdin: %w", err)
				}
			}

			oldTmpl, err := decoy.TemplateSvc.Get(name)
			if err != nil {
				if !errors.Is(err, errs.ErrNotFound) {
					return err
				}
			} else {
				fmt.Printf("Updating template %q:\n%s\n", name, oldTmpl)
				entity, err := decoy.TemplateSvc.Update(name, tmpl)
				if err != nil {
					return err
				}
				tmpl = entity.Tmpl
				fmt.Printf("to:\n%s\n", tmpl)
				return nil
			}

			entity, err := decoy.TemplateSvc.Save(name, tmpl)
			fmt.Printf("Saving new template %q:\n%s\n", name, entity.Tmpl)

			return err
		},
	}

	command.Flags().StringVarP(&tmpl, "template", "t", "", "The content of the template")
	command.Flags().StringVarP(&file, "file", "f", "", "The path of the file with the template contents")
	command.MarkFlagsMutuallyExclusive("template", "file")

	return command
}
