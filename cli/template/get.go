package template

import (
	"fmt"

	"github.com/aaron70/decoy/internal/services"
	"github.com/spf13/cobra"
)

func createGetCommand(decoy *services.Decoy) *cobra.Command {
	command := &cobra.Command{
		Use:   "get [<name>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "List all the templates or get the contents of the given template.",
		Example: `# List all templates
decoy template get

# Get the contents of the given template
decoy template get "greet"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				entity, err := decoy.TemplateSvc.Get(args[0])
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", entity.Tmpl)
			} else {
				entities, err := decoy.TemplateSvc.GetAll()
				if err != nil {
					return err
				}
				for _, entity := range entities {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", entity.Name)
				}
			}
			return nil
		},
	}
	return command
}
