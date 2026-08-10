package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/domehahn/skcr/v2/internal/skillversion"
	"github.com/spf13/cobra"
)

func newHistoryCommand() *cobra.Command {
	var repoPath string
	var skillFilter string
	var all bool
	var since string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "history [skill-dir]",
		Short: "Show version history timeline for one or all skills",
		Long: `Reads CHANGELOG.md entries and displays a structured version timeline.
Pass a skill directory as argument, or use --all to show all skills.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			type entry struct {
				Skill   string `json:"skill"`
				Version string `json:"version"`
				Date    string `json:"date"`
				Change  string `json:"change"`
			}

			var skillDirs []string

			if len(args) == 1 {
				abs, err := filepath.Abs(args[0])
				if err != nil {
					return err
				}
				skillDirs = []string{abs}
			} else {
				absTarget, err := filepath.Abs(repoPath)
				if err != nil {
					return err
				}
				cfg, loadErr := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
				if loadErr != nil {
					return loadErr
				}
				sourceBase := skillSourceOutputDir(cfg.SkillSources)

				seen := map[string]struct{}{}
				var names []string
				for _, t := range cfg.Targets {
					for _, s := range t.Skills {
						if _, dup := seen[s]; !dup {
							seen[s] = struct{}{}
							names = append(names, s)
						}
					}
				}
				sort.Strings(names)

				if skillFilter != "" {
					names = []string{skillFilter}
				} else if !all {
					return fmt.Errorf("provide a skill directory argument, --skill <name>, or --all")
				}

				for _, name := range names {
					skillDirs = append(skillDirs, filepath.Join(absTarget, sourceBase, name))
				}
			}

			var allEntries []entry
			for _, dir := range skillDirs {
				releases, err := skillversion.Changelog(dir)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "skip %s: %v\n", filepath.Base(dir), err)
					continue
				}
				for _, r := range releases {
					if since != "" && r.Date < since {
						continue
					}
					allEntries = append(allEntries, entry{
						Skill:   r.Name,
						Version: r.Version,
						Date:    r.Date,
						Change:  r.Change,
					})
				}
			}

			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(allEntries)
			}

			if len(allEntries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No changelog entries found.")
				return nil
			}

			prevSkill := ""
			for _, e := range allEntries {
				if e.Skill != prevSkill {
					if prevSkill != "" {
						fmt.Fprintln(cmd.OutOrStdout())
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n%s\n", e.Skill, strings.Repeat("─", 60))
					prevSkill = e.Skill
				}
				change := e.Change
				if len(change) > 72 {
					change = change[:69] + "..."
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  v%-10s  %-12s  %s\n", e.Version, e.Date, change)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path (used with --all or --skill)")
	cmd.Flags().StringVar(&skillFilter, "skill", "", "Show history for this skill only")
	cmd.Flags().BoolVar(&all, "all", false, "Show history for all skills in the bakefile")
	cmd.Flags().StringVar(&since, "since", "", "Only show entries on or after this date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	_ = cmd.RegisterFlagCompletionFunc("skill", completeSkillNames)
	return cmd
}
