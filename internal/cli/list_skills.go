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
)

func newListSkillsCommand() *cobra.Command {
	var target string
	var inTarget string
	var withTargets bool
	var withVersions bool
	var notScaffolded bool
	var orphaned bool
	var since string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "skills",
		Short: "List all skills defined across bakefile targets",
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := filepath.Abs(target)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}

			sourceBase := skillSourceOutputDir(cfg.SkillSources)

			// skill name → sorted list of target names that contain it
			skillTargets := map[string][]string{}

			for tn, t := range cfg.Targets {
				if inTarget != "" && tn != inTarget {
					continue
				}
				for _, s := range t.Skills {
					skillTargets[s] = append(skillTargets[s], tn)
				}
			}

			if len(skillTargets) == 0 {
				if inTarget != "" {
					fmt.Printf("No skills defined in target %q.\n", inTarget)
				} else {
					fmt.Println("No skills defined in any target.")
				}
				return nil
			}

			names := make([]string, 0, len(skillTargets))
			for name := range skillTargets {
				names = append(names, name)
			}
			sort.Strings(names)

			type skillEntry struct {
				Name      string   `json:"name"`
				Version   string   `json:"version,omitempty"`
				Targets   []string `json:"targets,omitempty"`
				Scaffolded bool    `json:"scaffolded"`
			}

			// --orphaned: skills on disk not in any bakefile target.
			if orphaned {
				return listOrphanedSkills(cmd, absTarget, sourceBase, skillTargets, jsonOut)
			}

			// --since: parse threshold once before iterating.
			var sinceTime time.Time
			if since != "" {
				var parseErr error
				sinceTime, parseErr = time.Parse("2006-01-02", since)
				if parseErr != nil {
					return fmt.Errorf("--since: expected YYYY-MM-DD, got %q", since)
				}
			}

			var entries []skillEntry
			for _, name := range names {
				scaffolded := true
				skillMD := filepath.Join(absTarget, sourceBase, name, "SKILL.md")
				if _, statErr := os.Stat(skillMD); statErr != nil {
					scaffolded = false
				}
				if notScaffolded && scaffolded {
					continue
				}

				// --since: filter by last_modified date in frontmatter.
				if !sinceTime.IsZero() && scaffolded {
					data, readErr := os.ReadFile(skillMD)
					if readErr != nil {
						continue
					}
					fm, ok := parseFrontmatterDates(string(data))
					if !ok || fm.LastModified == "" {
						continue
					}
					lm, parseErr := time.Parse("2006-01-02", fm.LastModified)
					if parseErr != nil || lm.Before(sinceTime) {
						continue
					}
				}

				targets := append([]string(nil), skillTargets[name]...)
				sort.Strings(targets)

				entry := skillEntry{
					Name:       name,
					Scaffolded: scaffolded,
					Targets:    targets,
				}
				if withVersions || jsonOut {
					entry.Version = skillVersion(absTarget, sourceBase, name)
				}
				entries = append(entries, entry)
			}

			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(entries)
			}

			for _, e := range entries {
				line := fmt.Sprintf("%-40s", e.Name)
				if withVersions {
					line += fmt.Sprintf("  %-12s", e.Version)
				}
				if withTargets {
					line += "  " + strings.Join(e.Targets, ", ")
				}
				fmt.Fprintln(cmd.OutOrStdout(), strings.TrimRight(line, " "))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Repository path")
	cmd.Flags().StringVar(&inTarget, "in-target", "", "List skills only from this bake target")
	cmd.Flags().BoolVar(&withTargets, "with-targets", false, "Show which bake targets each skill belongs to")
	cmd.Flags().BoolVar(&withVersions, "with-versions", false, "Show current version from SKILL.md")
	cmd.Flags().BoolVar(&notScaffolded, "not-scaffolded", false, "Show only skills without a SKILL.md on disk")
	cmd.Flags().BoolVar(&orphaned, "orphaned", false, "Show only skills present on disk but not in any bakefile target")
	cmd.Flags().StringVar(&since, "since", "", "Show only skills with last_modified on or after this date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	_ = cmd.RegisterFlagCompletionFunc("in-target", completeBakeTargets)
	return cmd
}

// listOrphanedSkills prints skill directories that exist on disk but are not
// referenced in any bakefile target.
func listOrphanedSkills(cmd *cobra.Command, absTarget, sourceBase string, inBakefile map[string][]string, jsonOut bool) error {
	baseDir := filepath.Join(absTarget, sourceBase)
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(cmd.OutOrStdout(), "No skill source directory found.")
			return nil
		}
		return err
	}

	type orphanEntry struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}
	var orphans []orphanEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		// Only count directories that contain a SKILL.md (real skill dirs).
		if _, statErr := os.Stat(filepath.Join(baseDir, name, "SKILL.md")); os.IsNotExist(statErr) {
			continue
		}
		if _, inBake := inBakefile[name]; !inBake {
			orphans = append(orphans, orphanEntry{Name: name, Path: filepath.Join(sourceBase, name)})
		}
	}

	if len(orphans) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No orphaned skills found.")
		return nil
	}

	if jsonOut {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(orphans)
	}
	for _, o := range orphans {
		fmt.Fprintf(cmd.OutOrStdout(), "%-40s  %s\n", o.Name, o.Path)
	}
	return nil
}

// skillVersion reads the version field from a skill's SKILL.md frontmatter.
// Returns "-" when the file is absent or has no version.
func skillVersion(absTarget, sourceBase, name string) string {
	path := filepath.Join(absTarget, sourceBase, name, "SKILL.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return "-"
	}
	fm, ok := parseFrontmatterDates(string(data))
	if !ok || fm.Version == "" {
		return "-"
	}
	return fm.Version
}
