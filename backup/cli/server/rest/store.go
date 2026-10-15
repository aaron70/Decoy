package rest

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
		name, spec, file string
		err              error
	)
	command := &cobra.Command{
		Use:   "store <name>",
		Short: "Upserts a server.",
		Long:  "Upserts a server. You can pass the OpenAPI Spec from stdin, a file, or through the spec flag.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name = args[0]

			if cmd.Flags().Changed("file") {
				bytes, err := os.ReadFile(file)
				if err != nil {
					return err
				}
				spec = strings.TrimRight(string(bytes), "\n")
			} else if !cmd.Flags().Changed("spec") {
				spec, err = utils.ReadStringFrom(cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("Couldn't read the config from stdin: %w", err)
				}
			}

			_, err := decoy.ServerSvc.Get(name)
			if err != nil {
				if !errors.Is(err, errs.ErrNotFound) {
					return err
				}
			} else {
				updated, err := decoy.ServerSvc.Update(name, spec)
				if err != nil {
					return err
				}
				fmt.Printf("Server %s has been updated\n", updated.Name)
				return nil
			}

			entity, err := decoy.ServerSvc.Save(name, spec)
			if err != nil {
				return err
			}

			fmt.Printf("Saving new server %s\n", entity.Name)
			return nil
		},
	}

	command.Flags().StringVarP(&spec, "spec", "s", "", "The contents of the OpenAPI Spec v3")
	command.Flags().StringVarP(&file, "file", "f", "", "The path of the file with the OpenAPI Spec v3")
	command.MarkFlagsMutuallyExclusive("spec", "file")

	return command
}
