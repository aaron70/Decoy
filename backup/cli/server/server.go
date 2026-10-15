package server

import (
	"github.com/aaron70/decoy/cli/server/rest"
	"github.com/aaron70/decoy/internal/services"
	"github.com/spf13/cobra"
)

func CreateServerCommand(decoy *services.Decoy) *cobra.Command {
	cmd := &cobra.Command{
		Use: "server",
		Short: "Manages the Decoy servers",
	}

	cmd.AddCommand(rest.CreateRestCommand(decoy))

	return cmd
}
