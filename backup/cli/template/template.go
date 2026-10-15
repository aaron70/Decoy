package template

import (
	"github.com/aaron70/decoy/internal/services"
	"github.com/spf13/cobra"
)

func CreateTemplateCommand(decoy *services.Decoy) *cobra.Command {
	command := &cobra.Command{
		Use:   "template",
		Short: "Groups the template commands",
		Long: `Manages the template commands, you can store, get, delete, and parse templates.
The Template engine uses Go's text/template library, so you can leverage Go's template syntax to structure and generate your data.`,
	}

	command.AddCommand(createStoreCommand(decoy))
	command.AddCommand(createGetCommand(decoy))
	command.AddCommand(createDeleteCommand(decoy))
	command.AddCommand(CreateParseCommand(decoy))

	return command
}
