package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/domehahn/skcr/internal/skillversion"
	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print skcr version information or manage skill versions",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "skcr %s (commit %s, built %s)\n", Version, Commit, Date)
		},
	}
	cmd.AddCommand(newVersionCheckCommand())
	cmd.AddCommand(newVersionBumpCommand())
	cmd.AddCommand(newVersionChangedCommand())
	cmd.AddCommand(newVersionChangelogCommand())
	cmd.AddCommand(newVersionReleaseNotesCommand())
	cmd.AddCommand(newVersionReleaseBundleCommand())
	cmd.AddCommand(newVersionSyncCommand())
	cmd.AddCommand(newVersionTagCommand())
	return cmd
}

func newVersionTagCommand() *cobra.Command {
	var dryRun bool
	var push bool
	var prefix string

	cmd := &cobra.Command{
		Use:               "tag <skill-dir>",
		Short:             "Create a git tag for the current skill version",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeSkillNames,
		RunE: func(cmd *cobra.Command, args []string) error {
			absPath, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}
			data, err := os.ReadFile(filepath.Join(absPath, "SKILL.md"))
			if err != nil {
				return fmt.Errorf("cannot read SKILL.md: %w", err)
			}
			fm, ok := parseFrontmatterDates(string(data))
			if !ok || fm.Version == "" {
				return fmt.Errorf("no version found in SKILL.md frontmatter")
			}
			name := filepath.Base(absPath)
			tagName := prefix + name + "/v" + fm.Version

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "would create tag  %s\n", tagName)
				return nil
			}

			if out, err := exec.Command("git", "tag", tagName).CombinedOutput(); err != nil {
				return fmt.Errorf("git tag: %s", strings.TrimSpace(string(out)))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created tag  %s\n", tagName)

			if push {
				if out, err := exec.Command("git", "push", "origin", tagName).CombinedOutput(); err != nil {
					return fmt.Errorf("git push: %s", strings.TrimSpace(string(out)))
				}
				fmt.Fprintf(cmd.OutOrStdout(), "pushed  %s\n", tagName)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show tag without creating it")
	cmd.Flags().BoolVar(&push, "push", false, "Push the tag to origin after creating it")
	cmd.Flags().StringVar(&prefix, "prefix", "skills/", "Tag name prefix")
	return cmd
}

func newVersionSyncCommand() *cobra.Command {
	var target string
	var dryRun bool
	var skillFilter string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Propagate SKILL.md frontmatter version to VERSION and skill.yaml for all skills",
		Long: `Reads each skill's SKILL.md frontmatter and writes the version into the
accompanying VERSION and skill.yaml files without incrementing it.
Useful after manual frontmatter edits.`,
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

			if len(skills) == 0 {
				fmt.Println("No skills defined in any target.")
				return nil
			}

			synced, skipped, errs := 0, 0, []string{}
			for _, name := range skills {
				skillDir := filepath.Join(absTarget, sourceBase, name)
				if _, statErr := os.Stat(filepath.Join(skillDir, "SKILL.md")); os.IsNotExist(statErr) {
					fmt.Printf("skip  %-40s  (no SKILL.md)\n", name)
					skipped++
					continue
				}
				if dryRun {
					fmt.Printf("would sync  %s/%s\n", sourceBase, name)
					synced++
					continue
				}
				info, syncErr := skillversion.SyncArtifacts(skillDir)
				if syncErr != nil {
					errs = append(errs, fmt.Sprintf("%s: %v", name, syncErr))
					continue
				}
				fmt.Printf("synced  %-40s  %s\n", name, info.Version)
				synced++
			}

			verb := "Sync complete"
			if dryRun {
				verb = "Dry run"
			}
			fmt.Printf("\n%s: %d synced, %d skipped.\n", verb, synced, skipped)
			if len(errs) > 0 {
				fmt.Fprintf(os.Stderr, "Errors:\n  %s\n", strings.Join(errs, "\n  "))
				return fmt.Errorf("%d skill(s) failed to sync", len(errs))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing")
	cmd.Flags().StringVar(&skillFilter, "skill", "", "Sync only this skill")
	_ = cmd.RegisterFlagCompletionFunc("skill", completeSkillNames)
	return cmd
}

func newVersionCheckCommand() *cobra.Command {
	var jsonOut bool
	var changed bool
	var all bool
	var target string
	cmd := &cobra.Command{
		Use:   "check [path]",
		Short: "Check SKILL.md version metadata",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var paths []string
			if all {
				absTarget, err := filepath.Abs(target)
				if err != nil {
					return err
				}
				cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
				if err != nil {
					return err
				}
				sourceBase := skillSourceOutputDir(cfg.SkillSources)
				seen := map[string]struct{}{}
				for _, t := range cfg.Targets {
					for _, s := range t.Skills {
						if _, dup := seen[s]; !dup {
							seen[s] = struct{}{}
							paths = append(paths, filepath.Join(absTarget, sourceBase, s))
						}
					}
				}
				sort.Strings(paths)
			} else if len(args) == 1 {
				paths = []string{args[0]}
			} else {
				return fmt.Errorf("provide a path or use --all")
			}

			var allInfos []skillversion.SkillInfo
			var allChanged []skillversion.ChangedSkill
			for _, p := range paths {
				infos, err := skillversion.Check(p)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "error checking %s: %v\n", p, err)
					continue
				}
				allInfos = append(allInfos, infos...)
				if changed {
					c, err := skillversion.Changed(p)
					if err == nil {
						allChanged = append(allChanged, c...)
					}
				}
			}

			if jsonOut {
				return writeJSON(cmd, map[string]any{"checks": allInfos, "changed": allChanged})
			}
			failed := false
			for _, info := range allInfos {
				status := "ok"
				if len(info.Errors) > 0 {
					status = "error"
					failed = true
				} else if len(info.Warnings) > 0 {
					status = "warn"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n", status, info.Name, info.Version, info.Path)
				for _, warning := range info.Warnings {
					fmt.Fprintf(cmd.OutOrStdout(), "WARN\t%s\t%s\n", info.Name, warning)
				}
				for _, e := range info.Errors {
					fmt.Fprintf(cmd.OutOrStdout(), "ERROR\t%s\t%s\n", info.Name, e)
				}
			}
			for _, item := range allChanged {
				status := "changed"
				if len(item.Errors) > 0 {
					status = "error"
					failed = true
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n", status, item.Name, item.CurrentVersion, item.Path)
				for _, e := range item.Errors {
					fmt.Fprintf(cmd.OutOrStdout(), "ERROR\t%s\t%s\n", item.Name, e)
				}
			}
			if failed {
				return fmt.Errorf("skill version check failed")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	cmd.Flags().BoolVar(&changed, "changed", false, "Check changed skills against git HEAD and require a version bump")
	cmd.Flags().BoolVar(&all, "all", false, "Check all skills defined in the bakefile")
	cmd.Flags().StringVarP(&target, "target", "t", ".", "Repository path (used with --all)")
	return cmd
}

func newVersionBumpCommand() *cobra.Command {
	var kind string
	var date string
	var change string
	var dryRun bool
	var preview bool
	var allChanged bool
	var allSkills bool
	var interactive bool
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "bump <skill-dir-or-bakefile-dir>",
		Short: "Bump a skill version and synchronize changelogs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if interactive {
				return runInteractiveBump(cmd, args[0], kind, date, dryRun)
			}
			if change == "" {
				return fmt.Errorf("required flag \"change\" not set (or use --interactive)")
			}
			if preview {
				dryRun = true
			}
			opts := skillversion.BumpOptions{Kind: skillversion.BumpKind(kind), Date: date, Change: change, DryRun: dryRun}
			if allChanged {
				results, err := skillversion.BumpAllChanged(args[0], opts)
				if err != nil {
					return err
				}
				if jsonOut {
					return writeJSON(cmd, results)
				}
				for _, result := range results {
					verb := "bumped"
					if result.DryRun {
						verb = "would-bump"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n", verb, result.Info.Name, result.NewVersion, result.Info.Path)
				}
				return nil
			}
			if allSkills {
				absBase, err := filepath.Abs(args[0])
				if err != nil {
					return err
				}
				cfg, err := cliLoadBakeFile(filepath.Join(absBase, "agentic.bake.yaml"))
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
				var results []skillversion.BumpResult
				for _, name := range names {
					skillDir := filepath.Join(absBase, sourceBase, name)
					if _, statErr := os.Stat(filepath.Join(skillDir, "SKILL.md")); os.IsNotExist(statErr) {
						continue
					}
					result, bumpErr := skillversion.BumpWithOptions(skillDir, opts)
					if bumpErr != nil {
						fmt.Fprintf(cmd.ErrOrStderr(), "skip %s: %v\n", name, bumpErr)
						continue
					}
					results = append(results, result)
				}
				if jsonOut {
					return writeJSON(cmd, results)
				}
				for _, result := range results {
					verb := "bumped"
					if result.DryRun {
						verb = "would-bump"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n", verb, result.Info.Name, result.NewVersion, result.Info.Path)
				}
				return nil
			}
			result, err := skillversion.BumpWithOptions(args[0], opts)
			if err != nil {
				return err
			}
			if jsonOut {
				return writeJSON(cmd, result)
			}
			if preview {
				dateStr := date
				if dateStr == "" {
					dateStr = "today"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Skill:    %s\n", result.Info.Name)
				fmt.Fprintf(cmd.OutOrStdout(), "Path:     %s\n", result.Info.Path)
				fmt.Fprintf(cmd.OutOrStdout(), "Current:  %s\n", result.OldVersion)
				fmt.Fprintf(cmd.OutOrStdout(), "New:      %s\n", result.NewVersion)
				fmt.Fprintf(cmd.OutOrStdout(), "Kind:     %s\n", kind)
				fmt.Fprintf(cmd.OutOrStdout(), "\nWould update:\n")
				fmt.Fprintf(cmd.OutOrStdout(), "  SKILL.md    (version: %s → %s, last_modified updated)\n", result.OldVersion, result.NewVersion)
				fmt.Fprintf(cmd.OutOrStdout(), "  VERSION     (%s → %s)\n", result.OldVersion, result.NewVersion)
				fmt.Fprintf(cmd.OutOrStdout(), "  skill.yaml  (version: %s → %s)\n", result.OldVersion, result.NewVersion)
				fmt.Fprintf(cmd.OutOrStdout(), "  CHANGELOG.md (new entry: [%s] %s — %s)\n", result.NewVersion, dateStr, change)
				return nil
			}
			verb := "bumped"
			if result.DryRun {
				verb = "would-bump"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n", verb, result.Info.Name, result.NewVersion, result.Info.Path)
			return nil
		},
	}
	cmd.Flags().StringVar(&kind, "kind", "patch", "Version bump kind: major, minor, or patch")
	cmd.Flags().StringVar(&date, "date", "", "Changelog date in YYYY-MM-DD format (default: today)")
	cmd.Flags().StringVar(&change, "change", "", "Machine-readable changelog message")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing files")
	cmd.Flags().BoolVar(&preview, "preview", false, "Show a detailed human-readable preview without writing (implies --dry-run)")
	cmd.Flags().BoolVar(&allChanged, "all-changed", false, "Bump every changed skill that has not already changed version")
	cmd.Flags().BoolVar(&allSkills, "all", false, "Bump all scaffolded skills defined in the bakefile")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Prompt for bump kind and message per changed skill")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	return cmd
}

func runInteractiveBump(cmd *cobra.Command, path, defaultKind, date string, dryRun bool) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	changed, err := skillversion.Changed(absPath)
	if err != nil {
		// Not inside a git repo or git unavailable — treat as no changes.
		fmt.Fprintf(cmd.OutOrStdout(), "No changed skills detected (git: %v).\n", err)
		return nil
	}
	if len(changed) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No changed skills detected.")
		return nil
	}

	reader := bufio.NewReader(cmd.InOrStdin())
	prompt := func(label, def string) (string, error) {
		if def != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s [%s]: ", label, def)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s: ", label)
		}
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			return def, nil
		}
		return line, nil
	}

	var results []skillversion.BumpResult
	for _, item := range changed {
		fmt.Fprintf(cmd.OutOrStdout(), "\n─── %s  (current: %s)\n", item.Name, item.CurrentVersion)
		kindVal, err := prompt("kind (major/minor/patch/skip)", defaultKind)
		if err != nil {
			return err
		}
		if kindVal == "skip" || kindVal == "s" {
			fmt.Fprintf(cmd.OutOrStdout(), "  skipped\n")
			continue
		}
		changeMsg, err := prompt("changelog message", "")
		if err != nil {
			return err
		}
		if changeMsg == "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  skipped (no message)\n")
			continue
		}
		opts := skillversion.BumpOptions{
			Kind:   skillversion.BumpKind(kindVal),
			Date:   date,
			Change: changeMsg,
			DryRun: dryRun,
		}
		result, bumpErr := skillversion.BumpWithOptions(item.Path, opts)
		if bumpErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "  error: %v\n", bumpErr)
			continue
		}
		verb := "bumped"
		if dryRun {
			verb = "would-bump"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  %s → %s\n", verb, result.NewVersion)
		results = append(results, result)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "\nDone: %d skill(s) bumped.\n", len(results))
	return nil
}

func newVersionChangedCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "changed <path>",
		Short: "List changed skills and whether their version changed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			changed, err := skillversion.Changed(args[0])
			if err != nil {
				return err
			}
			if jsonOut {
				return writeJSON(cmd, changed)
			}
			failed := false
			for _, item := range changed {
				status := "changed"
				if len(item.Errors) > 0 {
					status = "error"
					failed = true
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n", status, item.Name, item.CurrentVersion, item.Path)
				for _, file := range item.Files {
					fmt.Fprintf(cmd.OutOrStdout(), "FILE\t%s\t%s\n", item.Name, file)
				}
				for _, e := range item.Errors {
					fmt.Fprintf(cmd.OutOrStdout(), "ERROR\t%s\t%s\n", item.Name, e)
				}
			}
			if failed {
				return fmt.Errorf("changed skill without version bump")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	return cmd
}

func newVersionChangelogCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "changelog <path>",
		Short: "Print machine-readable skill changelog entries",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := skillversion.Changelog(args[0])
			if err != nil {
				return err
			}
			if jsonOut {
				return writeJSON(cmd, entries)
			}
			for _, entry := range entries {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n", entry.Date, entry.Name, entry.Version, entry.Change)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	return cmd
}

func newVersionReleaseNotesCommand() *cobra.Command {
	var since string
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "release-notes <path>",
		Short: "Generate release notes from skill changelogs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			notes, err := skillversion.ReleaseNotes(args[0], since)
			if err != nil {
				return err
			}
			if jsonOut {
				return writeJSON(cmd, map[string]string{"release_notes": notes})
			}
			fmt.Fprint(cmd.OutOrStdout(), strings.TrimRight(notes, "\n")+"\n")
			return nil
		},
	}
	cmd.Flags().StringVar(&since, "since", "", "Only include changelog entries on or after this YYYY-MM-DD date")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	return cmd
}

func newVersionReleaseBundleCommand() *cobra.Command {
	var since string
	var changed bool
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "release-bundle <path>",
		Short: "Generate a release bundle with checks, changelog, and release notes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundle, err := skillversion.ReleaseBundleFor(args[0], since, changed)
			if err != nil {
				return err
			}
			if jsonOut {
				return writeJSON(cmd, bundle)
			}
			fmt.Fprint(cmd.OutOrStdout(), strings.TrimRight(bundle.ReleaseNotes, "\n")+"\n")
			fmt.Fprintf(cmd.OutOrStdout(), "\nChecks: %d\nChangelog entries: %d\n", len(bundle.Checks), len(bundle.Changelog))
			if changed {
				fmt.Fprintf(cmd.OutOrStdout(), "Changed skills: %d\n", len(bundle.Changed))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&since, "since", "", "Only include changelog entries on or after this YYYY-MM-DD date")
	cmd.Flags().BoolVar(&changed, "changed", false, "Include git changed-skill report when available")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	return cmd
}

func writeJSON(cmd *cobra.Command, value any) error {
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
