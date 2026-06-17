package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newCheckCompatCommand() *cobra.Command {
	var target string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "check-compat <skill> <platform>",
		Short: "Verify a skill's frontmatter lists a specific platform",
		Long: `Reads the platforms: list from <skill>/SKILL.md and exits non-zero if
<platform> is not present. Useful as a CI gate before baking to a specific
target platform.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName, platform := args[0], args[1]

			if err := validateSkillArg(skillName); err != nil {
				return err
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
			skillMD := filepath.Join(abs, sourceBase, skillName, "SKILL.md")

			data, err := cliReadFile(skillMD)
			if err != nil {
				return fmt.Errorf("read SKILL.md: %w", err)
			}

			platforms := parseFrontmatterSlice(string(data), "platforms")

			found := false
			for _, p := range platforms {
				if strings.EqualFold(p, platform) {
					found = true
					break
				}
			}

			if jsonOut {
				if err := json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
					"skill":      skillName,
					"platform":   platform,
					"compatible": found,
					"platforms":  platforms,
				}); err != nil {
					return err
				}
				if !found {
					return fmt.Errorf("%s does not list platform %q", skillName, platform)
				}
				return nil
			}

			if found {
				fmt.Fprintf(cmd.OutOrStdout(), "%s is compatible with %s\n", skillName, platform)
				return nil
			}
			return fmt.Errorf("%s does not list platform %q (platforms: %s)", skillName, platform, strings.Join(platforms, ", "))
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}
