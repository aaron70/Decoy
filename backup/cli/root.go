package cli

import (
	"os"
	"path/filepath"

	"github.com/aaron70/decoy/cli/runner"
	"github.com/aaron70/decoy/cli/server"
	"github.com/aaron70/decoy/cli/server/rest"
	"github.com/aaron70/decoy/cli/template"
	"github.com/aaron70/decoy/internal/services"
	"github.com/spf13/cobra"
)

func createRootCommand(decoy *services.Decoy) *cobra.Command {
	command := &cobra.Command{
		Use:   "decoy",
		Short: "A CLI tool to generate mock data",
		Long:  "Decoy is a command-line utility for creating and ingesting mock data through templates and runners. Templates support dynamic data injection using the Go Template Engine, and runners let you ingest the generated data into your application.",
	}

	command.AddCommand(
		template.CreateTemplateCommand(decoy),
		template.CreateParseCommand(decoy),
		runner.CreateRunnerCommand(decoy),
		runner.CreateRunCommand(decoy),
		server.CreateServerCommand(decoy),
		rest.CreateStartCommand(decoy),
	)

	return command
}

func configDir() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "decoy")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("/tmp", ".config", "decoy")
	}
	return filepath.Join(home, ".config", "decoy")
}

func Execute() error {
	cli, err := services.NewDecoyFS(configDir())
	if err != nil {
		return err
	}
	defer cli.Decoy.Close()
	err = createRootCommand(cli).Execute()
	if err != nil {
		return err
	}
	return nil
}
