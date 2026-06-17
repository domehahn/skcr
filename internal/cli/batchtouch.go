package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newBatchTouchCommand() *cobra.Command {
	var target string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "batch-touch",
		Short: "Update last_modified on skills whose files are newer than their recorded date",
		Long: `Walks every skill directory under skill_sources.output_dir, reads the
last_modified: field from SKILL.md frontmatter, and updates it to today if the
file's mtime is more recent than the recorded date.

Use after bulk edits to bring last_modified dates in sync.`,
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
				return fmt.Errorf("read skill sources dir: %w", err)
			}

			today := time.Now().UTC().Format("2006-01-02")
			updated, skipped := 0, 0

			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				skillMD := filepath.Join(sourceDir, entry.Name(), "SKILL.md")
				fi, err := os.Stat(skillMD)
				if err != nil {
					continue
				}

				raw, err := cliReadFile(skillMD)
				if err != nil {
					continue
				}

				// Extract recorded last_modified from frontmatter via proper YAML parse.
				// YAML may parse a bare date as time.Time whose Sprintf form is long —
				// truncate to YYYY-MM-DD for a reliable lexicographic comparison.
				recorded := parseFrontmatterScalar(string(raw), "last_modified")
				if len(recorded) > 10 {
					recorded = recorded[:10]
				}
				mtime := fi.ModTime().UTC().Format("2006-01-02")

				if recorded >= mtime {
					skipped++
					continue
				}

				if !strings.HasPrefix(string(raw), "---\n") {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s has no frontmatter — skipping\n",
						filepath.Join(sourceBase, entry.Name(), "SKILL.md"))
					skipped++
					continue
				}
				newContent, changed := setFrontmatterField(string(raw), "last_modified", today)
				if !changed {
					skipped++
					continue
				}

				rel := filepath.Join(sourceBase, entry.Name(), "SKILL.md")
				if dryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "  would touch  %s  (was %s, mtime %s)\n", rel, recorded, mtime)
					updated++
					continue
				}

				if err := cliWriteFile(skillMD, []byte(newContent), 0o644); err != nil {
					return fmt.Errorf("write %s: %w", skillMD, err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  touched  %s\n", rel)
				updated++
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "\nDry run: would touch %d skill(s), skip %d.\n", updated, skipped)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nTouched %d skill(s), skipped %d (already up to date).\n", updated, skipped)
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	return cmd
}
