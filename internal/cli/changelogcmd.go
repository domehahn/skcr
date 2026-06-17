package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/domehahn/skcr/internal/skillversion"
	"github.com/spf13/cobra"
)

func newChangelogCommand() *cobra.Command {
	var repoPath string
	var since string
	var bakeTarget string
	var outPath string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "changelog",
		Short: "Aggregate CHANGELOG.md entries across all skills into a combined document",
		Long: `Reads each skill's CHANGELOG.md (via skillversion.Changelog) and merges all
entries into a single release notes document, sorted newest-first.

Use --since <YYYY-MM-DD> to include only entries on or after that date.
Use --bake-target to restrict to one bakefile target (default: all targets).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return err
			}
			sourceBase := skillSourceOutputDir(cfg.SkillSources)

			seen := map[string]struct{}{}
			var skills []string
			for tName, t := range cfg.Targets {
				if bakeTarget != "" && tName != bakeTarget {
					continue
				}
				for _, s := range t.Skills {
					if _, dup := seen[s]; !dup {
						seen[s] = struct{}{}
						skills = append(skills, s)
					}
				}
			}
			sort.Strings(skills)

			var all []skillversion.ReleaseEntry
			for _, skillName := range skills {
				entries, _ := skillversion.Changelog(filepath.Join(abs, sourceBase, skillName))
				for _, e := range entries {
					if since != "" && e.Date < since {
						continue
					}
					if e.Name == "" {
						e.Name = skillName
					}
					all = append(all, e)
				}
			}

			// Sort: newest date first, then by skill name.
			sort.Slice(all, func(i, j int) bool {
				if all[i].Date != all[j].Date {
					return all[i].Date > all[j].Date
				}
				return all[i].Name < all[j].Name
			})

			if jsonOut {
				data, err := json.MarshalIndent(all, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal: %w", err)
				}
				return writeOrPrint(cmd, outPath, string(data)+"\n")
			}

			var sb strings.Builder
			sb.WriteString("# Combined Changelog\n\n")
			if since != "" {
				sb.WriteString(fmt.Sprintf("_Since %s_\n\n", since))
			}
			if len(all) == 0 {
				sb.WriteString("No changelog entries found.\n")
			}
			for _, e := range all {
				sb.WriteString(fmt.Sprintf("## %s — %s@%s\n\n- %s\n\n", e.Date, e.Name, e.Version, e.Change))
			}

			return writeOrPrint(cmd, outPath, sb.String())
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().StringVar(&since, "since", "", "Only include entries on or after this date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&bakeTarget, "bake-target", "", "Restrict to a single bakefile target")
	cmd.Flags().StringVar(&outPath, "out", "", "Write output to file instead of stdout")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON array")
	return cmd
}

func writeOrPrint(cmd *cobra.Command, path, content string) error {
	if path == "" {
		fmt.Fprint(cmd.OutOrStdout(), content)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
