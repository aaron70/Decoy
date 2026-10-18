package runner

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
		name, config, file string
		interactive        bool
	)
	command := &cobra.Command{
		Use:   "store <name>",
		Short: "Upserts a runner.",
		Long:  "Upserts a runner. You can pass the config JSON from stdin, a file, or through the config flag.",
		Example: `# Store a runner config from stdin
echo 'echo "{{ .template }}"' | decoy runner store "echo"

# Store a runner config from a file
decoy runner store "echo" -f /path/to/config

# Store an inline runner config
decoy runner store "echo" -c 'echo "{{ .template }}"'
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name = args[0]

			isUpdate := true
			oldRunner, err := decoy.RunnerSvc.Get(name)
			if err != nil {
				if !errors.Is(err, errs.ErrNotFound) {
					return err
				}
				isUpdate = false
			}

			if interactive {
				initialText := ""
				if isUpdate {
					initialText = oldRunner.Config
				}
				bytes, err := utils.EditText(initialText)
				if err != nil {
					return err
				}
				config = string(bytes)
			} else if cmd.Flags().Changed("file") {
				bytes, err := os.ReadFile(file)
				if err != nil {
					return err
				}
				config = strings.TrimRight(string(bytes), "\n")
			} else if !cmd.Flags().Changed("config") {
				config, err = utils.ReadStringFrom(cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("Couldn't read the config from stdin: %w", err)
				}
			}

			if isUpdate {
				fmt.Printf("Updating runner %q:\n%s\n", name, string(oldRunner.Config))
				entity, err := decoy.RunnerSvc.Update(name, config)
				if err != nil {
					return err
				}
				config = string(entity.Config)
				fmt.Printf("to:\n%s\n", config)
				return nil
			}

			entity, err := decoy.RunnerSvc.Save(name, config)
			fmt.Printf("Saving new runner %q:\n%s\n", name, string(entity.Config))

			return err
		},
	}

	command.Flags().StringVarP(&config, "config", "c", "", "The config JSON of the runner")
	command.Flags().StringVarP(&file, "file", "f", "", "The path of the file with the runner config")
	command.Flags().BoolVarP(&interactive, "interactive", "i", false, "Whether to open an editor to edit the the runner.")
	command.MarkFlagsMutuallyExclusive("config", "file")

	return command
}
