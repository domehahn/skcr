package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

func newWatchCommand() *cobra.Command {
	var repoPath string
	var intervalSec int
	var runFmt bool
	var runBake bool

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch SKILL.md files and re-run validate on changes",
		Long: `Polls skill directories for SKILL.md changes and automatically runs
validate (and optionally fmt or bake --plan) on the changed skill.
Press Ctrl-C to stop.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}
			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
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

			// Build initial mtime snapshot.
			mtimes := map[string]time.Time{}
			for _, name := range names {
				p := filepath.Join(absTarget, sourceBase, name, "SKILL.md")
				if info, statErr := os.Stat(p); statErr == nil {
					mtimes[p] = info.ModTime()
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Watching %d skill(s) — press Ctrl-C to stop.\n", len(names))

			ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
			defer ticker.Stop()

			exec := func(subArgs ...string) {
				root := NewRootCommand()
				root.SetArgs(subArgs)
				root.SetOut(cmd.OutOrStdout())
				root.SetErr(cmd.ErrOrStderr())
				if runErr := root.Execute(); runErr != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "  [error] %v\n", runErr)
				}
			}

			for range ticker.C {
				// Reload skill list in case bakefile changed.
				freshCfg, reloadErr := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
				if reloadErr == nil {
					seen = map[string]struct{}{}
					for _, t := range freshCfg.Targets {
						for _, s := range t.Skills {
							seen[s] = struct{}{}
						}
					}
				}

				for name := range seen {
					p := filepath.Join(absTarget, sourceBase, name, "SKILL.md")
					info, statErr := os.Stat(p)
					if statErr != nil {
						continue
					}
					prev, known := mtimes[p]
					if known && info.ModTime().Equal(prev) {
						continue
					}
					mtimes[p] = info.ModTime()
					if !known {
						fmt.Fprintf(cmd.OutOrStdout(), "[new]     %s\n", name)
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "[changed] %s  %s\n", name, info.ModTime().Format("15:04:05"))
					}

					if runFmt {
						fmt.Fprintf(cmd.OutOrStdout(), "  → fmt\n")
						exec("fmt", "--target", absTarget, "--skill", name)
					}
					fmt.Fprintf(cmd.OutOrStdout(), "  → validate\n")
					exec("validate", "--target", absTarget)
					if runBake {
						fmt.Fprintf(cmd.OutOrStdout(), "  → bake --plan\n")
						exec("bake", "--target", absTarget, "--plan")
					}
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().IntVar(&intervalSec, "interval", 1, "Poll interval in seconds")
	cmd.Flags().BoolVar(&runFmt, "fmt", false, "Run fmt on changed skills before validate")
	cmd.Flags().BoolVar(&runBake, "bake", false, "Run bake --plan after validate")
	return cmd
}
