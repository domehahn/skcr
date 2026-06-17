package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const skcrSnapshotDir = ".skcr-snapshots"

type skillSnapshot struct {
	Name         string `json:"name" yaml:"name"`
	Version      string `json:"version" yaml:"version"`
	Since        string `json:"since,omitempty" yaml:"since,omitempty"`
	LastModified string `json:"last_modified,omitempty" yaml:"last_modified,omitempty"`
	Path         string `json:"path" yaml:"path"`
}

type snapshotManifest struct {
	SavedAt string          `json:"saved_at" yaml:"saved_at"`
	Label   string          `json:"label" yaml:"label"`
	Skills  []skillSnapshot `json:"skills" yaml:"skills"`
}

func newSnapshotCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Checkpoint and restore SKILL.md version metadata before bulk edits",
		Long: `Saves and restores the version/last_modified/since fields of all SKILL.md files
in the skill sources. Useful before a bulk refactor when you want a rollback
point that is independent of git state.`,
	}
	cmd.AddCommand(newSnapshotSaveCommand())
	cmd.AddCommand(newSnapshotRestoreCommand())
	cmd.AddCommand(newSnapshotListCommand())
	cmd.AddCommand(newSnapshotDeleteCommand())
	return cmd
}

func newSnapshotSaveCommand() *cobra.Command {
	var repoPath string
	var label string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "save <name>",
		Short: "Save current SKILL.md version metadata as a named snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			abs, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return err
			}
			sourceBase := skillSourceOutputDir(cfg.SkillSources)

			var skills []skillSnapshot
			seen := map[string]struct{}{}
			for _, t := range cfg.Targets {
				for _, skillName := range t.Skills {
					if _, dup := seen[skillName]; dup {
						continue
					}
					seen[skillName] = struct{}{}
					skillMD := filepath.Join(abs, sourceBase, skillName, "SKILL.md")
					data, readErr := os.ReadFile(skillMD)
					if readErr != nil {
						continue
					}
					fm, ok := parseFrontmatterDates(string(data))
					if !ok {
						continue
					}
					skills = append(skills, skillSnapshot{
						Name:         skillName,
						Version:      fm.Version,
						Since:        fm.Since,
						LastModified: fm.LastModified,
						Path:         skillMD,
					})
				}
			}
			sort.Slice(skills, func(i, j int) bool { return skills[i].Name < skills[j].Name })

			snap := snapshotManifest{
				SavedAt: time.Now().UTC().Format(time.RFC3339),
				Label:   label,
				Skills:  skills,
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Dry run: would save %d skill(s) as snapshot %q\n", len(skills), name)
				return nil
			}

			snapDir := filepath.Join(abs, skcrSnapshotDir)
			if err := os.MkdirAll(snapDir, 0o755); err != nil {
				return fmt.Errorf("create snapshot dir: %w", err)
			}
			snapPath := filepath.Join(snapDir, name+".json")
			data, err := json.MarshalIndent(snap, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal snapshot: %w", err)
			}
			if err := os.WriteFile(snapPath, data, 0o644); err != nil {
				return fmt.Errorf("write snapshot: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Snapshot %q saved: %d skill(s) → %s\n", name, len(skills), snapPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().StringVar(&label, "label", "", "Optional human-readable label")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	return cmd
}

func newSnapshotRestoreCommand() *cobra.Command {
	var repoPath string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "restore <name>",
		Short: "Restore SKILL.md version metadata from a named snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			abs, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}
			snapPath := filepath.Join(abs, skcrSnapshotDir, name+".json")
			data, err := os.ReadFile(snapPath)
			if err != nil {
				return fmt.Errorf("snapshot %q not found — run: skcr snapshot list", name)
			}

			var snap snapshotManifest
			if err := json.Unmarshal(data, &snap); err != nil {
				return fmt.Errorf("parse snapshot: %w", err)
			}

			restored := 0
			for _, ss := range snap.Skills {
				skillPath := ss.Path
				if !filepath.IsAbs(skillPath) {
					skillPath = filepath.Join(abs, skillPath)
				}
				content, readErr := os.ReadFile(skillPath)
				if readErr != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "  !  skip %s: not found\n", ss.Name)
					continue
				}

				updated, changed := restoreFrontmatterFields(string(content), ss)
				if !changed {
					fmt.Fprintf(cmd.OutOrStdout(), "  -  %s: no change\n", ss.Name)
					continue
				}

				if dryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "  ~  %s: would restore to %s\n", ss.Name, ss.Version)
					continue
				}

				if err := os.WriteFile(skillPath, []byte(updated), 0o644); err != nil {
					return fmt.Errorf("write %s: %w", skillPath, err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  ✓  %s: restored to %s\n", ss.Name, ss.Version)
				restored++
			}

			if !dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "\nRestored %d of %d skill(s) from snapshot %q (saved %s)\n",
					restored, len(snap.Skills), name, snap.SavedAt)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	return cmd
}

func newSnapshotListCommand() *cobra.Command {
	var repoPath string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available snapshots",
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}
			snapDir := filepath.Join(abs, skcrSnapshotDir)
			entries, err := os.ReadDir(snapDir)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Fprintln(cmd.OutOrStdout(), "No snapshots found.")
					return nil
				}
				return fmt.Errorf("read snapshot dir: %w", err)
			}

			found := 0
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
					continue
				}
				snapName := strings.TrimSuffix(e.Name(), ".json")
				data, err := os.ReadFile(filepath.Join(snapDir, e.Name()))
				if err != nil {
					continue
				}
				var snap snapshotManifest
				if json.Unmarshal(data, &snap) != nil {
					continue
				}
				label := snap.Label
				if label != "" {
					label = "  " + label
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  %-24s  %s  (%d skills)%s\n",
					snapName, snap.SavedAt, len(snap.Skills), label)
				found++
			}
			if found == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No snapshots found.")
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	return cmd
}

func newSnapshotDeleteCommand() *cobra.Command {
	var repoPath string
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a named snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			abs, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}
			snapPath := filepath.Join(abs, skcrSnapshotDir, name+".json")
			if err := os.Remove(snapPath); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("snapshot %q not found", name)
				}
				return fmt.Errorf("delete snapshot: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted snapshot %q\n", name)
			return nil
		},
	}
	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	return cmd
}

// restoreFrontmatterFields rewrites version, last_modified, since in SKILL.md frontmatter.
func restoreFrontmatterFields(content string, ss skillSnapshot) (string, bool) {
	if !strings.HasPrefix(content, "---\n") {
		return content, false
	}
	end := strings.Index(content[4:], "\n---")
	if end < 0 {
		return content, false
	}
	fmRaw := content[4 : 4+end]
	var fm map[string]any
	if err := yaml.Unmarshal([]byte(fmRaw), &fm); err != nil {
		return content, false
	}

	changed := false
	if ss.Version != "" && fm["version"] != ss.Version {
		fm["version"] = ss.Version
		changed = true
	}
	if ss.LastModified != "" && fm["last_modified"] != ss.LastModified {
		fm["last_modified"] = ss.LastModified
		changed = true
	}
	if ss.Since != "" && fm["since"] != ss.Since {
		fm["since"] = ss.Since
		changed = true
	}
	if !changed {
		return content, false
	}

	newFM, err := yaml.Marshal(fm)
	if err != nil {
		return content, false
	}
	body := content[4+end+4:]
	return "---\n" + string(newFM) + "---" + body, true
}
