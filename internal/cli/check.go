package cli

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newCheckCommand() *cobra.Command {
	var repoPath string
	var failOnWarn bool
	var skipFmt bool
	var skipLint bool
	var skipAudit bool

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Pre-commit gate: fmt + lint + audit in one shot",
		Long: `Runs all quality checks in sequence and reports a single pass/fail result.
Ideal for a git pre-commit hook or local CI step.

Steps (all run against the same --target):
  1. fmt    (dry-run — reports drift without writing)
  2. lint   (content quality: placeholders, short descriptions)
  3. audit  (structural completeness: metadata, version drift)

Use --skip-fmt / --skip-lint / --skip-audit to skip individual steps.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			type step struct {
				name string
				args []string
				skip bool
			}

			steps := []step{
				{"fmt", []string{"fmt", "--dry-run", "--target", repoPath}, skipFmt},
				{"lint", buildLintArgs(repoPath, failOnWarn), skipLint},
				{"audit", []string{"audit", "--target", repoPath}, skipAudit},
			}

			type result struct {
				name   string
				passed bool
				output string
				err    error
			}

			var results []result
			anyFailed := false

			for _, s := range steps {
				if s.skip {
					results = append(results, result{name: s.name, passed: true, output: "(skipped)"})
					continue
				}

				var buf bytes.Buffer
				root := NewRootCommand()
				root.SetArgs(s.args)
				root.SetOut(&buf)
				root.SetErr(&buf)
				err := root.Execute()

				passed := err == nil
				if !passed {
					anyFailed = true
				}
				results = append(results, result{
					name:   s.name,
					passed: passed,
					output: strings.TrimSpace(buf.String()),
					err:    err,
				})
			}

			fmt.Fprintln(cmd.OutOrStdout())
			for _, r := range results {
				icon := "✓"
				if !r.passed {
					icon = "✗"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  %s  %s\n", icon, r.name)
				if r.output != "" && r.output != "(skipped)" {
					for _, line := range strings.Split(r.output, "\n") {
						fmt.Fprintf(cmd.OutOrStdout(), "       %s\n", line)
					}
				}
			}
			fmt.Fprintln(cmd.OutOrStdout())

			if anyFailed {
				fmt.Fprintf(cmd.OutOrStdout(), "check: FAILED — fix the issues above before committing\n")
				return fmt.Errorf("check failed")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "check: all steps passed\n")
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&failOnWarn, "fail-on-warn", false, "Treat lint warnings as failures")
	cmd.Flags().BoolVar(&skipFmt, "skip-fmt", false, "Skip the fmt step")
	cmd.Flags().BoolVar(&skipLint, "skip-lint", false, "Skip the lint step")
	cmd.Flags().BoolVar(&skipAudit, "skip-audit", false, "Skip the audit step")
	return cmd
}

func buildLintArgs(repoPath string, failOnWarn bool) []string {
	a := []string{"lint", "--target", repoPath}
	if failOnWarn {
		a = append(a, "--fail-on-warn")
	}
	return a
}
