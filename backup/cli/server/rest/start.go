package rest

import (
	"bytes"
	"io"
	"os"

	"github.com/aaron70/decoy/api/rest"
	"github.com/aaron70/decoy/internal/services"
	"github.com/aaron70/decoy/internal/utils"
	"github.com/spf13/cobra"
)

func CreateStartCommand(decoy *services.Decoy) *cobra.Command {
	var (
		noSpec     bool
		specFile   io.ReadCloser
		err        error
		file, spec string
	)
	cmd := &cobra.Command{
		Use:   "start [<name>]",
		Short: "Starts a new server to listen for requests",
		Long:  `Starts a new server to listen for request. The decoy server allows you to manage the decoy resources through the /api endpoint or mock a server through the /mock endpoint.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !noSpec {
				if len(args) > 0 {
					server, err := decoy.ServerSvc.Get(args[0])
					if err != nil { return err }
					spec = server.Spec
				} else if cmd.Flags().Changed("file") {
					specFile, err = os.OpenFile(file, os.O_RDONLY, os.ModePerm)
					if err != nil {
						return err
					}
				} else if !cmd.Flags().Changed("spec") {
					spec, err = utils.ReadStringFrom(cmd.InOrStdin())
					if err != nil {
						return err
					}
				}

				if specFile == nil {
					specFile = io.NopCloser(bytes.NewBufferString(spec))
				}
			}

			return rest.Start(cmd.OutOrStdout(), decoy, 8080, specFile)
		},
	}

	cmd.Flags().BoolVar(&noSpec, "no-spec", false, "Whether the server should mock a OpenAPI Spec or not")
	cmd.Flags().StringVarP(&file, "file", "f", "", "The path to the OpenAPI Spec v3 file")
	cmd.Flags().StringVarP(&spec, "spec", "s", "", "The contents of the OpenAPI Spec v3")
	cmd.MarkFlagsMutuallyExclusive("no-spec", "file", "spec")

	return cmd
}
