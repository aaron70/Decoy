package rest

import (
	"github.com/aaron70/decoy/internal/services"
	"github.com/spf13/cobra"
)

func CreateRestCommand(decoy *services.Decoy) *cobra.Command {
	cmd := &cobra.Command{
		Use: "rest",
		Short: "Manages servers of type REST",
	}

	cmd.AddCommand(createStoreCommand(decoy))
	cmd.AddCommand(CreateStartCommand(decoy))
	cmd.AddCommand(createDeleteCommand(decoy))
	cmd.AddCommand(createGetCommand(decoy))

	return cmd
}
