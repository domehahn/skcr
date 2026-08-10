package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/domehahn/skcr/v2/internal/catalog"
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
	var scope string
	cmd := &cobra.Command{
		Use:   "categories",
		Short: "List built-in skill categories",
		RunE: func(cmd *cobra.Command, args []string) error {
			categories, err := categoryNamesForScope(scope)
			if err != nil {
				return err
			}
			for _, category := range categories {
				fmt.Fprintf(cmd.OutOrStdout(), "%-28s  %d skill(s)\n", category, len(catalog.SkillCategories[category]))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&scope, "scope", "all", "category scope: all, semantic, or cncf")
	return cmd
}

func categoryNamesForScope(scope string) ([]string, error) {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "all":
		return catalog.CategoryNames(), nil
	case "semantic":
		return catalog.SemanticCategoryNames(), nil
	case "cncf":
		var names []string
		for _, name := range catalog.CategoryNames() {
			if name == "cncf" || strings.HasPrefix(name, "cncf-") {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		return names, nil
	default:
		return nil, fmt.Errorf("unknown category scope %q (available: all, semantic, cncf)", scope)
	}
}
