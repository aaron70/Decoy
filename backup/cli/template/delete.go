package template

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
		Short: "Deletes the given template.",
		Example: `# Delete a template
decoy template delete "greet"

# Delete all templates without auto-confirmation
decoy template delete --all

# Delete all templates with auto-confirmation
decoy template delete --all --yes`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) <= 0 && deleteAll {

				if !confirmed {
					res, err := utils.AskForInput(cmd.InOrStdin(), cmd.OutOrStdout(), "Are you sure you want to delete all templates? (y/n): ")
					if err != nil {
						return err
					}
					if strings.ToLower(res) != "y" && strings.ToLower(res) != "yes" {
						fmt.Fprintf(cmd.OutOrStdout(), "Action canceled by the user!\n")
						return nil
					}
				}

				templates, err := decoy.TemplateSvc.GetAll()
				if err != nil {
					return err
				}

				for _, tmpl := range templates {
					_, err := decoy.TemplateSvc.Delete(tmpl.Name)
					if err != nil {
						return err
					}
				}

				fmt.Fprintf(cmd.OutOrStdout(), "All templates have been deleted\n")
				return nil
			} else if len(args) <= 0 {
				return errors.New("Missing the name of the template to delete, if you want to delete all templates use the --all/-a flag and confirm the operation with --yes/-y flag.")
			} else {
				name = args[0]
			}

			entity, err := decoy.TemplateSvc.Delete(name)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Template %q has been deleted:\n%s\n", entity.Name, entity.Tmpl)
			return nil
		},
	}

	command.Flags().BoolVarP(&deleteAll, "all", "a", false, "Delete all saved templates")
	command.Flags().BoolVarP(&confirmed, "yes", "y", false, "Confirms whether you are sure to delete all templates")

	return command
}
