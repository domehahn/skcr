package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func newSyncCommand() *cobra.Command {
	var target string
	var dryRun bool
	var skillFilter string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Propagate canonical skill artifacts to all platform directories",
		Long: `Reads canonical artifacts from .agents/skills/<name>/ and copies SKILL.md,
skill.yaml, contract.yaml, and evals/ to every other platform skill directory
(.claude/skills/, .github/skills/, etc.)
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

			sourceBase := skillSourceOutputDir(cfg.SkillSources)
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
				srcSkillDir := filepath.Join(absTarget, sourceBase, s)
				srcPath := filepath.Join(srcSkillDir, "SKILL.md")
				_, err := os.ReadFile(srcPath)
				if err != nil {
					fmt.Printf("skip  %-40s  (no .agents/skills/%s/SKILL.md)\n", s, s)
					skipped++
					continue
				}

				for _, destDir := range destDirs {
					destSkillDir := filepath.Join(absTarget, destDir, s)
					if _, statErr := cliStatBake(destSkillDir); os.IsNotExist(statErr) {
						continue // not scaffolded yet; bake --write will create it
					}
					artifacts, walkErr := canonicalSkillArtifacts(srcSkillDir)
					if walkErr != nil {
						return walkErr
					}
					for _, rel := range artifacts {
						srcData, readErr := os.ReadFile(filepath.Join(srcSkillDir, rel))
						if readErr != nil {
							return readErr
						}
						destPath := filepath.Join(destSkillDir, rel)
						existing, _ := os.ReadFile(destPath)
						if string(existing) == string(srcData) {
							unchanged++
							continue
						}
						if dryRun {
							fmt.Printf("would update  %s/%s/%s\n", destDir, s, filepath.ToSlash(rel))
							updated++
							continue
						}
						if err := rejectSymlinkArtifactPath(destSkillDir, destPath); err != nil {
							return err
						}
						if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
							return err
						}
						if err := os.WriteFile(destPath, srcData, 0o644); err != nil {
							return fmt.Errorf("sync %s/%s/%s: %w", destDir, s, filepath.ToSlash(rel), err)
						}
						fmt.Printf("updated  %s/%s/%s\n", destDir, s, filepath.ToSlash(rel))
						updated++
					}
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

func canonicalSkillArtifacts(skillDir string) ([]string, error) {
	artifacts := []string{}
	for _, rel := range []string{"SKILL.md", "contract.yaml", "skill.yaml"} {
		info, err := os.Lstat(filepath.Join(skillDir, rel))
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("canonical artifact must not be a symlink: %s", filepath.Join(skillDir, rel))
		} else if err == nil {
			artifacts = append(artifacts, rel)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	evalsDir := filepath.Join(skillDir, "evals")
	err := filepath.WalkDir(evalsDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) && path == evalsDir {
				return nil
			}
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("canonical artifact must not be a symlink: %s", path)
		}
		rel, err := filepath.Rel(skillDir, path)
		if err != nil {
			return err
		}
		artifacts = append(artifacts, rel)
		return nil
	})
	sort.Strings(artifacts)
	return artifacts, err
}

func rejectSymlinkArtifactPath(skillDir, destination string) error {
	rel, err := filepath.Rel(skillDir, destination)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("artifact destination escapes skill directory: %s", destination)
	}
	current := skillDir
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		if info, statErr := os.Lstat(current); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("artifact destination must not traverse symlink: %s", current)
		} else if statErr != nil && !os.IsNotExist(statErr) {
			return statErr
		}
	}
	return nil
}
