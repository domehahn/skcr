package cli

import (
	"fmt"

	"github.com/domehahn/skcr/internal/catalog"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List bakefile resources (skills, targets)",
	}
	cmd.AddCommand(newListSkillsCommand())
	cmd.AddCommand(newListTargetsSubCommand())
	cmd.AddCommand(newListCategoriesCommand())
	return cmd
}

// newListTargetsSubCommand wraps list-targets as a subcommand of "list".
func newListTargetsSubCommand() *cobra.Command {
	cmd := newListTargetsCommand()
	cmd.Use = "targets"
	return cmd
}

func newListCategoriesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "categories",
		Short: "List built-in skill categories",
		Run: func(cmd *cobra.Command, args []string) {
			for _, category := range catalog.CategoryNames() {
				fmt.Fprintf(cmd.OutOrStdout(), "%-28s  %d skill(s)\n", category, len(catalog.SkillCategories[category]))
			}
		},
	}
}
