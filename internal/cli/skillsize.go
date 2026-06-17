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

type skillSizeEntry struct {
	Skill string `json:"skill"`
	Bytes int64  `json:"bytes"`
	Lines int    `json:"lines"`
}

func newSkillSizeCommand() *cobra.Command {
	var target string
	var sortBy string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "skill-size",
		Short: "Report byte size and line count for each skill's SKILL.md",
		Long: `Walks every skill directory under skill_sources.output_dir and reports the
size of SKILL.md in bytes and lines. Use --sort to order the results.

Useful for identifying skills that have grown too large and should be split.`,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			var rows []skillSizeEntry
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				skillMD := filepath.Join(sourceDir, e.Name(), "SKILL.md")
				fi, err := os.Stat(skillMD)
				if err != nil {
					continue
				}
				data, err := os.ReadFile(skillMD)
				if err != nil {
					continue
				}
				lines := strings.Count(string(data), "\n")
				if len(data) > 0 && data[len(data)-1] != '\n' {
					lines++
				}
				rows = append(rows, skillSizeEntry{
					Skill: e.Name(),
					Bytes: fi.Size(),
					Lines: lines,
				})
			}

			switch sortBy {
			case "name":
				sort.Slice(rows, func(i, j int) bool { return rows[i].Skill < rows[j].Skill })
			case "size":
				sort.Slice(rows, func(i, j int) bool { return rows[i].Bytes > rows[j].Bytes })
			case "lines":
				sort.Slice(rows, func(i, j int) bool { return rows[i].Lines > rows[j].Lines })
			default:
				return fmt.Errorf("unknown sort %q: must be name, size, or lines", sortBy)
			}

			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(rows)
			}

			if len(rows) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No skills found.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-32s  %8s  %6s\n", "SKILL", "BYTES", "LINES")
			fmt.Fprintln(cmd.OutOrStdout(), strings.Repeat("-", 52))
			for _, r := range rows {
				fmt.Fprintf(cmd.OutOrStdout(), "%-32s  %8d  %6d\n", r.Skill, r.Bytes, r.Lines)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().StringVar(&sortBy, "sort", "name", "Sort order: name, size, or lines")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}
