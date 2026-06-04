package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/agentic-template-kit/skcr/internal/bake"
	"github.com/spf13/cobra"
)

func newListTargetsCommand() *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "list-targets",
		Short: "List available bake targets",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := filepath.Join(target, "agentic.bake.yaml")
			cfg, err := bake.LoadBakeFile(path)
			if err != nil {
				return err
			}

			names := make([]string, 0, len(cfg.Targets))
			for name := range cfg.Targets {
				names = append(names, name)
			}
			sort.Strings(names)

			fmt.Println("Target\tDescription\tPlatforms / Inherits")
			for _, name := range names {
				t := cfg.Targets[name]
				value := strings.Join(t.Platforms, ", ")
				if value == "" {
					value = strings.Join(t.Inherits, ", ")
				}
				fmt.Printf("%s\t%s\t%s\n", name, t.Description, value)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Target repository path")
	return cmd
}
