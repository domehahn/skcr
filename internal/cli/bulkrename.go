package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newBulkRenameCommand() *cobra.Command {
	var target string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "bulk-rename <old-prefix> <new-prefix>",
		Short: "Rename all skills matching a prefix, updating SKILL.md and bakefile references",
		Long: `Finds every skill directory under skill_sources.output_dir whose name starts
with <old-prefix>, renames the directory, updates name: in SKILL.md, and rewrites
all target and flow references in agentic.bake.yaml.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldPrefix, newPrefix := args[0], args[1]
			if oldPrefix == newPrefix {
				fmt.Fprintln(cmd.OutOrStdout(), "Prefixes are identical — nothing to do.")
				return nil
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

			type rename struct{ oldName, newName string }
			var renames []rename
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				if strings.HasPrefix(e.Name(), oldPrefix) {
					newName := newPrefix + strings.TrimPrefix(e.Name(), oldPrefix)
					if err := validateSkillArg(newName); err != nil {
						return fmt.Errorf("computed name %q is invalid: %w", newName, err)
					}
					renames = append(renames, rename{e.Name(), newName})
				}
			}

			if len(renames) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No skills with prefix %q found.\n", oldPrefix)
				return nil
			}

			for _, r := range renames {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s → %s\n", r.oldName, r.newName)
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "\nDry run: would rename %d skill(s).\n", len(renames))
				return nil
			}

			// Pre-check: none of the destinations may already exist.
			for _, r := range renames {
				if _, err := os.Stat(filepath.Join(sourceDir, r.newName)); err == nil {
					return fmt.Errorf("destination %q already exists — aborting", r.newName)
				}
			}

			// Apply renames.
			for _, r := range renames {
				oldPath := filepath.Join(sourceDir, r.oldName)
				newPath := filepath.Join(sourceDir, r.newName)

				// Update name: in SKILL.md.
				skillMD := filepath.Join(oldPath, "SKILL.md")
				if data, err := cliReadFile(skillMD); err == nil {
					updated, _ := setFrontmatterField(string(data), "name", r.newName)
					if err := cliWriteFile(skillMD, []byte(updated), 0o644); err != nil {
						return fmt.Errorf("update SKILL.md for %s: %w", r.oldName, err)
					}
				}

				if err := os.Rename(oldPath, newPath); err != nil {
					return fmt.Errorf("rename %s → %s: %w", r.oldName, r.newName, err)
				}
			}

			// Build rename map for O(1) lookup.
			renameMap := make(map[string]string, len(renames))
			for _, r := range renames {
				renameMap[r.oldName] = r.newName
			}

			// Rewrite bakefile by updating the data model, then dumping.
			for tName, tc := range cfg.Targets {
				if tc == nil {
					continue
				}
				updated := make([]string, len(tc.Skills))
				for i, s := range tc.Skills {
					if n, ok := renameMap[s]; ok {
						updated[i] = n
					} else {
						updated[i] = s
					}
				}
				cfg.Targets[tName].Skills = updated

				// Flow directory names are not skill names — do not rename them.
			}

			if err := cliDumpBakeFile(cfg, filepath.Join(abs, "agentic.bake.yaml")); err != nil {
				return fmt.Errorf("write bakefile: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\nRenamed %d skill(s) and updated agentic.bake.yaml.\n", len(renames))
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making changes")
	return cmd
}
