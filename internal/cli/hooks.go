package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const (
	hookBeginMarker = "# skcr:begin"
	hookEndMarker   = "# skcr:end"
	hookScript      = `
# skcr:begin — managed by skcr hooks install (do not edit this block)
if command -v skcr >/dev/null 2>&1; then
  skcr check || exit 1
fi
# skcr:end`
)

func newHooksCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hooks",
		Short: "Manage git hooks that run skcr checks automatically",
	}
	cmd.AddCommand(newHooksInstallCommand())
	cmd.AddCommand(newHooksUninstallCommand())
	cmd.AddCommand(newHooksStatusCommand())
	return cmd
}

func findGitDir(start string) (string, error) {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return filepath.Join(dir, ".git"), nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not inside a git repository (no .git found from %s)", start)
		}
		dir = parent
	}
}

func newHooksInstallCommand() *cobra.Command {
	var repoPath string
	var hookName string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install skcr check as a git pre-commit hook",
		Long: `Appends a managed block to .git/hooks/pre-commit that runs 'skcr check'
before every commit. The block is idempotent — running install twice is safe.

The hook is wrapped in a command-exists guard so repos without skcr installed
can still commit (useful when only some contributors have skcr).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}
			gitDir, err := findGitDir(abs)
			if err != nil {
				return err
			}

			hookPath := filepath.Join(gitDir, "hooks", hookName)

			var existing string
			if data, readErr := os.ReadFile(hookPath); readErr == nil {
				existing = string(data)
			}

			if strings.Contains(existing, hookBeginMarker) {
				fmt.Fprintf(cmd.OutOrStdout(), "Hook already installed: %s\n", hookPath)
				return nil
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Dry run: would append skcr block to %s\n", hookPath)
				return nil
			}

			if err := os.MkdirAll(filepath.Dir(hookPath), 0o755); err != nil {
				return fmt.Errorf("create hooks dir: %w", err)
			}

			content := existing
			if content == "" {
				content = "#!/bin/sh\n"
			}
			if !strings.HasSuffix(content, "\n") {
				content += "\n"
			}
			content += hookScript + "\n"

			if err := os.WriteFile(hookPath, []byte(content), 0o755); err != nil {
				return fmt.Errorf("write hook file: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Installed skcr pre-commit hook: %s\n", hookPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().StringVar(&hookName, "hook", "pre-commit", "Hook name (default: pre-commit)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	return cmd
}

func newHooksUninstallCommand() *cobra.Command {
	var repoPath string
	var hookName string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the skcr managed block from a git hook",
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}
			gitDir, err := findGitDir(abs)
			if err != nil {
				return err
			}

			hookPath := filepath.Join(gitDir, "hooks", hookName)
			data, err := os.ReadFile(hookPath)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Fprintf(cmd.OutOrStdout(), "Hook not found: %s\n", hookPath)
					return nil
				}
				return fmt.Errorf("read hook file: %w", err)
			}

			content := string(data)
			if !strings.Contains(content, hookBeginMarker) {
				fmt.Fprintf(cmd.OutOrStdout(), "No skcr block found in %s\n", hookPath)
				return nil
			}

			cleaned := removeHookBlock(content)

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Dry run: would remove skcr block from %s\n", hookPath)
				return nil
			}

			if strings.TrimSpace(cleaned) == "#!/bin/sh" || strings.TrimSpace(cleaned) == "" {
				if err := os.Remove(hookPath); err != nil {
					return fmt.Errorf("remove empty hook: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Removed empty hook file: %s\n", hookPath)
				return nil
			}

			if err := os.WriteFile(hookPath, []byte(cleaned), 0o755); err != nil {
				return fmt.Errorf("write hook file: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed skcr block from %s\n", hookPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().StringVar(&hookName, "hook", "pre-commit", "Hook name (default: pre-commit)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	return cmd
}

func newHooksStatusCommand() *cobra.Command {
	var repoPath string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show whether the skcr hook block is installed",
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}
			gitDir, err := findGitDir(abs)
			if err != nil {
				return err
			}

			hookPath := filepath.Join(gitDir, "hooks", "pre-commit")
			data, err := os.ReadFile(hookPath)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Fprintf(cmd.OutOrStdout(), "  -  pre-commit hook: not installed\n")
					return nil
				}
				return fmt.Errorf("read hook file: %w", err)
			}

			if strings.Contains(string(data), hookBeginMarker) {
				fmt.Fprintf(cmd.OutOrStdout(), "  ✓  pre-commit hook: installed (%s)\n", hookPath)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "  !  pre-commit hook exists but skcr block not found (%s)\n", hookPath)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	return cmd
}

// removeHookBlock strips everything between hookBeginMarker and hookEndMarker (inclusive).
func removeHookBlock(content string) string {
	lines := strings.Split(content, "\n")
	var out []string
	inside := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), hookBeginMarker) {
			inside = true
			continue
		}
		if strings.TrimSpace(line) == hookEndMarker {
			inside = false
			continue
		}
		if !inside {
			out = append(out, line)
		}
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n") + "\n"
}
