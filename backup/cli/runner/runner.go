package runner

import (
	"github.com/aaron70/decoy/internal/services"
	"github.com/spf13/cobra"
)

func CreateRunnerCommand(decoy *services.Decoy) *cobra.Command {
	command := &cobra.Command{
		Use:   "runner",
		Short: "Groups the runner commands",
	}

	command.AddCommand(createStoreCommand(decoy))
	command.AddCommand(createGetCommand(decoy))
	command.AddCommand(createDeleteCommand(decoy))
	command.AddCommand(CreateRunCommand(decoy))

	return command
}
