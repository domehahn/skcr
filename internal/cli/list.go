package cli

import "github.com/spf13/cobra"

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List bakefile resources (skills, targets)",
	}
	cmd.AddCommand(newListSkillsCommand())
	cmd.AddCommand(newListTargetsSubCommand())
	return cmd
}

// newListTargetsSubCommand wraps list-targets as a subcommand of "list".
func newListTargetsSubCommand() *cobra.Command {
	cmd := newListTargetsCommand()
	cmd.Use = "targets"
	return cmd
}
