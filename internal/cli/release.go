package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/domehahn/skcr/internal/skillversion"
	"github.com/spf13/cobra"
)

func newReleaseCommand() *cobra.Command {
	var kind string
	var change string
	var date string
	var tagPrefix string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "release <skill-dir>",
		Short: "Orchestrate a full release: bump version, update CHANGELOG, print git-tag",
		Long: `Runs the complete release workflow for a single skill in one shot:
  1. Validate the skill dir and read current version
  2. Bump SKILL.md version + last_modified + changelog entry
  3. Sync VERSION and skill.yaml artifacts
  4. Print the git-tag command to run

Use --dry-run to preview what would change without writing files.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			absDir, err := filepath.Abs(args[0])
			if err != nil {
				return fmt.Errorf("resolve path: %w", err)
			}

			if date == "" {
				date = time.Now().UTC().Format("2006-01-02")
			}
			if kind == "" {
				return fmt.Errorf("--kind is required (patch, minor, major)")
			}
			if change == "" {
				return fmt.Errorf("--change is required")
			}

			opts := skillversion.BumpOptions{
				Kind:   skillversion.BumpKind(kind),
				Date:   date,
				Change: change,
				DryRun: dryRun,
			}

			result, err := skillversion.BumpWithOptions(absDir, opts)
			if err != nil {
				return err
			}

			tag := tagPrefix + result.NewVersion
			if result.Info.Name != "" && !strings.HasPrefix(tag, result.Info.Name) {
				tag = result.Info.Name + "/" + tagPrefix + result.NewVersion
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Dry run — no files written\n\n")
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Skill:    %s\n", result.Info.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "Version:  %s → %s\n", result.OldVersion, result.NewVersion)
			fmt.Fprintf(cmd.OutOrStdout(), "Date:     %s\n", date)
			fmt.Fprintf(cmd.OutOrStdout(), "Change:   %s\n", change)

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "\nWould update:\n")
				fmt.Fprintf(cmd.OutOrStdout(), "  SKILL.md    (version, last_modified, changelog)\n")
				fmt.Fprintf(cmd.OutOrStdout(), "  VERSION     (%s)\n", result.NewVersion)
				fmt.Fprintf(cmd.OutOrStdout(), "  skill.yaml  (version: %s)\n", result.NewVersion)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "\nUpdated:\n")
				fmt.Fprintf(cmd.OutOrStdout(), "  SKILL.md, VERSION, skill.yaml\n")
				fmt.Fprintf(cmd.OutOrStdout(), "\nNext step — tag the release:\n")
				fmt.Fprintf(cmd.OutOrStdout(), "  git add %s\n", absDir)
				fmt.Fprintf(cmd.OutOrStdout(), "  git commit -m \"release: %s %s\"\n", result.Info.Name, result.NewVersion)
				fmt.Fprintf(cmd.OutOrStdout(), "  git tag %s\n", tag)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&kind, "kind", "", "Bump kind: patch, minor, or major (required)")
	cmd.Flags().StringVar(&change, "change", "", "Changelog entry message (required)")
	cmd.Flags().StringVar(&date, "date", "", "Release date in YYYY-MM-DD format (default: today)")
	cmd.Flags().StringVar(&tagPrefix, "tag-prefix", "v", "Prefix for the git tag")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing files")
	return cmd
}
