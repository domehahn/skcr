package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func newFilterByCommand() *cobra.Command {
	var target string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "filter-by <key> <value>",
		Short: "List all skills where a frontmatter field matches a value",
		Long: `Walks every skill directory under skill_sources.output_dir, reads each
SKILL.md frontmatter, and prints skills where <key>: equals <value>.

Examples:
  skcr filter-by stability experimental
  skcr filter-by owner platform-team
  skcr filter-by version 1.0.0`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]

			if value == "" {
				return fmt.Errorf("value must not be empty")
			}

			abs, err := filepath.Abs(target)
			if err != nil {
				return fmt.Errorf("resolve target: %w", err)
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return fmt.Errorf("load bakefile: %w", err)
			}

			sourceBase := skillSourceOutputDir(cfg.SkillSources)
			sourceDir := filepath.Join(abs, sourceBase)

			entries, err := os.ReadDir(sourceDir)
			if err != nil {
				return fmt.Errorf("read skills dir: %w", err)
			}

			var matches []string
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				data, err := cliReadFile(filepath.Join(sourceDir, e.Name(), "SKILL.md"))
				if err != nil {
					continue
				}
				if strings.EqualFold(parseFrontmatterScalar(string(data), key), value) {
					matches = append(matches, e.Name())
				}
			}
			sort.Strings(matches)

			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
					"key":     key,
					"value":   value,
					"matches": matches,
					"count":   len(matches),
				})
			}

			if len(matches) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No skills with %s: %q\n", key, value)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%d skill(s) with %s: %q\n\n", len(matches), key, value)
			for _, m := range matches {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", m)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}
