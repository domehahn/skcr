package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

func newSyncCommand() *cobra.Command {
	var target string
	var dryRun bool
	var skillFilter string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Propagate SKILL.md edits from .agents/skills/ to all platform directories",
		Long: `Reads the canonical SKILL.md from .agents/skills/<name>/SKILL.md and copies it
to every other platform skill directory (.claude/skills/, .github/skills/, etc.)
where the skill is already scaffolded. Unscaffolded directories are skipped —
run "skcr bake --write" first to create them.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := filepath.Abs(target)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}

			resolved, err := resolveDefaultTarget(cfg)
			if err != nil {
				return err
			}

			const sourceBase = ".agents/skills"
			dirSeen := map[string]struct{}{sourceBase: {}}
			var destDirs []string
			for _, p := range resolved.Platforms {
				d := canonicalPlatformSkillBaseDir(p)
				if _, dup := dirSeen[d]; !dup {
					dirSeen[d] = struct{}{}
					destDirs = append(destDirs, d)
				}
			}

			if len(destDirs) == 0 {
				fmt.Println("No platform-specific skill directories configured.")
				return nil
			}

			// Collect all unique skill names from every target.
			skillSeen := map[string]struct{}{}
			var skills []string
			for _, t := range cfg.Targets {
				for _, s := range t.Skills {
					if _, dup := skillSeen[s]; !dup {
						skillSeen[s] = struct{}{}
						skills = append(skills, s)
					}
				}
			}
			sort.Strings(skills)

			if skillFilter != "" {
				skills = []string{skillFilter}
			}

			updated, unchanged, skipped := 0, 0, 0
			for _, s := range skills {
				srcPath := filepath.Join(absTarget, sourceBase, s, "SKILL.md")
				srcData, err := os.ReadFile(srcPath)
				if err != nil {
					fmt.Printf("skip  %-40s  (no .agents/skills/%s/SKILL.md)\n", s, s)
					skipped++
					continue
				}

				for _, destDir := range destDirs {
					destPath := filepath.Join(absTarget, destDir, s, "SKILL.md")
					if _, statErr := cliStatBake(destPath); os.IsNotExist(statErr) {
						continue // not scaffolded yet; bake --write will create it
					}
					existing, _ := os.ReadFile(destPath)
					if string(existing) == string(srcData) {
						unchanged++
						continue
					}
					if dryRun {
						fmt.Printf("would update  %s/%s/SKILL.md\n", destDir, s)
						updated++
						continue
					}
					if err := os.WriteFile(destPath, srcData, 0o644); err != nil {
						return fmt.Errorf("sync %s/%s/SKILL.md: %w", destDir, s, err)
					}
					fmt.Printf("updated  %s/%s/SKILL.md\n", destDir, s)
					updated++
				}
			}

			verb := "Sync complete"
			if dryRun {
				verb = "Dry run"
			}
			fmt.Printf("\n%s: %d updated, %d unchanged, %d missing canonical source.\n", verb, updated, unchanged, skipped)
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing")
	cmd.Flags().StringVar(&skillFilter, "skill", "", "Sync only this skill")
	_ = cmd.RegisterFlagCompletionFunc("skill", completeSkillNames)
	return cmd
}
