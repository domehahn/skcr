package cli

import (
	"encoding/json"
	"fmt"
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
	return cmd
}

func newVersionCheckCommand() *cobra.Command {
	var jsonOut bool
	var changed bool
	cmd := &cobra.Command{
		Use:   "check <path>",
		Short: "Check SKILL.md version metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			infos, err := skillversion.Check(args[0])
			if err != nil {
				return err
			}
			changedSkills := []skillversion.ChangedSkill{}
			if changed {
				changedSkills, err = skillversion.Changed(args[0])
				if err != nil {
					return err
				}
			}
			if jsonOut {
				return writeJSON(cmd, map[string]any{"checks": infos, "changed": changedSkills})
			}
			failed := false
			for _, info := range infos {
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
			for _, item := range changedSkills {
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
	return cmd
}

func newVersionBumpCommand() *cobra.Command {
	var kind string
	var date string
	var change string
	var dryRun bool
	var allChanged bool
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "bump <skill-dir-or-SKILL.md>",
		Short: "Bump a skill version and synchronize changelogs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			result, err := skillversion.BumpWithOptions(args[0], opts)
			if err != nil {
				return err
			}
			if jsonOut {
				return writeJSON(cmd, result)
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
	cmd.Flags().BoolVar(&allChanged, "all-changed", false, "Bump every changed skill that has not already changed version")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	_ = cmd.MarkFlagRequired("change")
	return cmd
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
