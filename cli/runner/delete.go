package runner

import (
	"fmt"
	"strings"

	"github.com/aaron70/decoy/internal/services"
	"github.com/aaron70/decoy/internal/utils"
	"github.com/aaron70/goaty/errors"
	"github.com/spf13/cobra"
)

func createDeleteCommand(decoy *services.Decoy) *cobra.Command {
	var (
		name      string
		deleteAll bool
		confirmed bool
	)
	command := &cobra.Command{
		Use:   "delete [<name>]",
		Short: "Deletes the given runner.",
		Example: `# Delete a runner
decoy runner delete "echo"

# Delete all runners without auto-confirmation
decoy runner delete --all

# Delete all runners with auto-confirmation
decoy runner delete --all --yes`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) <= 0 && deleteAll {

				if !confirmed {
					res, err := utils.AskForInput(cmd.InOrStdin(), cmd.OutOrStdout(), "Are you sure you want to delete all runners? (y/n): ")
					if err != nil {
						return err
					}
					if strings.ToLower(res) != "y" && strings.ToLower(res) != "yes" {
						fmt.Fprintf(cmd.OutOrStdout(), "Action canceled by the user!\n")
						return nil
					}
				}

				runners, err := decoy.RunnerSvc.GetAll()
				if err != nil {
					return err
				}

				for _, runner := range runners {
					_, err := decoy.RunnerSvc.Delete(runner.Name)
					if err != nil {
						return err
					}
				}

				fmt.Fprintf(cmd.OutOrStdout(), "All runners have been deleted\n")
				return nil
			} else if len(args) <= 0 {
				return errors.New("Missing the name of the runner to delete, if you want to delete all runners use the --all/-a flag and confirm the operation with --yes/-y flag.")
			} else {
				name = args[0]
			}

			entity, err := decoy.RunnerSvc.Delete(name)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Runner %q has been deleted:\n%s\n", entity.Name, string(entity.Config))
			return nil
		},
	}

	command.Flags().BoolVarP(&deleteAll, "all", "a", false, "Delete all saved runners")
	command.Flags().BoolVarP(&confirmed, "yes", "y", false, "Confirms whether you are sure to delete all runners")

	return command
}
