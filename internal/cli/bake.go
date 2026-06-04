package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/agentic-template-kit/skcr/internal/bake"
	"github.com/agentic-template-kit/skcr/internal/lockfile"
	"github.com/agentic-template-kit/skcr/internal/models"
	"github.com/agentic-template-kit/skcr/internal/renderer"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/cobra"
)

const maxDiffLinesPerFile = 120

var (
	cliAbsPathBake   = filepath.Abs
	cliLoadBakeFile  = bake.LoadBakeFile
	cliResolveTarget = bake.ResolveTarget
	cliRenderFiles   = renderer.RenderFiles
	cliLoadLockfile  = lockfile.LoadLockfile
	cliWriteLockfile = lockfile.WriteLockfile
	cliReadFile      = os.ReadFile
	cliMkdirAllBake  = os.MkdirAll
	cliWriteFile     = os.WriteFile
)

func newBakeCommand() *cobra.Command {
	var target string
	var plan bool
	var detailedDiff bool
	var write bool

	cmd := &cobra.Command{
		Use:   "bake [target]",
		Short: "Render files for a target and preview or write them",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetName := "default"
			if len(args) > 0 {
				targetName = args[0]
			}
			if !plan && !write {
				plan = true
			}

			absTarget, err := cliAbsPathBake(target)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}
			resolved, err := cliResolveTarget(cfg, targetName)
			if err != nil {
				return err
			}
			files, err := cliRenderFiles(cfg, resolved)
			if err != nil {
				return err
			}

			lock, err := cliLoadLockfile(absTarget)
			if err != nil {
				return err
			}
			stateFiles := lockfile.ManagedFilesByPath(lock)
			plannedFiles := map[string]models.RenderedFile{}
			for _, file := range files {
				plannedFiles[file.Destination] = file
			}

			fmt.Printf("Bake target: %s\n", targetName)
			fmt.Println("Action\tPlatform\tPath")
			for _, rendered := range sortedRendered(files) {
				path := filepath.Join(absTarget, rendered.Destination)
				action := "create"
				if current, err := cliReadFile(path); err == nil {
					if string(current) == rendered.Content {
						action = "unchanged"
					} else {
						action = "update"
					}
				}
				fmt.Printf("%s\t%s\t%s\n", action, rendered.Platform, rendered.Destination)
			}

			stateCounts := map[string]int{"create": 0, "update": 0, "delete": 0, "noop": 0}
			fmt.Println("\nState Plan (.agentic-template.lock)")
			fmt.Println("Action\tPlatform\tPath")

			for _, path := range sortedKeys(plannedFiles) {
				rendered := plannedFiles[path]
				checksum := lockfile.Sha256Text(rendered.Content)
				prev, ok := stateFiles[path]
				action := "create"
				if ok {
					if checksumValue(prev) == checksum {
						action = "noop"
					} else {
						action = "update"
					}
				}
				stateCounts[action]++
				fmt.Printf("%s\t%s\t%s\n", action, rendered.Platform, path)
			}
			for _, path := range sortedMapKeys(stateFiles) {
				if _, exists := plannedFiles[path]; exists {
					continue
				}
				stateCounts["delete"]++
				platform := "-"
				if p, ok := stateFiles[path]["platform"].(string); ok && p != "" {
					platform = p
				}
				fmt.Printf("delete\t%s\t%s\n", platform, path)
			}

			fmt.Printf("\nPlan summary: %d to create, %d to update, %d to delete, %d unchanged in state.\n", stateCounts["create"], stateCounts["update"], stateCounts["delete"], stateCounts["noop"])

			if plan {
				changed := []string{}
				for path, rendered := range plannedFiles {
					current, err := cliReadFile(filepath.Join(absTarget, path))
					if err != nil {
						continue
					}
					if string(current) != rendered.Content {
						changed = append(changed, path)
					}
				}
				sort.Strings(changed)
				for _, path := range changed {
					rendered := plannedFiles[path]
					current, err := cliReadFile(filepath.Join(absTarget, path))
					if err != nil {
						continue
					}
					diffText := unifiedDiff(string(current), rendered.Content, path, !detailedDiff)
					if strings.TrimSpace(diffText) == "" {
						continue
					}
					fmt.Printf("\nDiff: %s\n%s\n", path, diffText)
				}

				deleted := []string{}
				for path := range stateFiles {
					if _, ok := plannedFiles[path]; !ok {
						deleted = append(deleted, path)
					}
				}
				sort.Strings(deleted)
				if len(deleted) > 0 {
					fmt.Println("\nState-only files (would be removed from state if applied):")
					for _, path := range deleted {
						fmt.Println("-", path)
					}
				}

				fmt.Println("\nDry run only. Use --write to write files.")
				return nil
			}

			for _, rendered := range files {
				path := filepath.Join(absTarget, rendered.Destination)
				if err := cliMkdirAllBake(filepath.Dir(path), 0o755); err != nil {
					return err
				}
				if err := cliWriteFile(path, []byte(rendered.Content), 0o644); err != nil {
					return err
				}
			}

			if err := cliWriteLockfile(absTarget, files, targetName); err != nil {
				return err
			}
			fmt.Printf("Wrote %d files and .agentic-template.lock\n", len(files))
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Target repository path")
	cmd.Flags().BoolVar(&plan, "plan", false, "Show generated files without writing")
	cmd.Flags().BoolVar(&detailedDiff, "detailed-diff", false, "Show full unified diffs in --plan")
	cmd.Flags().BoolVar(&write, "write", false, "Write files to target repository")

	return cmd
}

func sortedRendered(files []models.RenderedFile) []models.RenderedFile {
	copyFiles := make([]models.RenderedFile, len(files))
	copy(copyFiles, files)
	sort.Slice(copyFiles, func(i, j int) bool {
		return copyFiles[i].Destination < copyFiles[j].Destination
	})
	return copyFiles
}

func sortedKeys(m map[string]models.RenderedFile) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedMapKeys(m map[string]map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func checksumValue(entry map[string]any) string {
	if value, ok := entry["checksum"].(string); ok {
		return value
	}
	return ""
}

func unifiedDiff(oldText, newText, path string, truncate bool) string {
	d := difflib.UnifiedDiff{
		A:        difflib.SplitLines(oldText),
		B:        difflib.SplitLines(newText),
		FromFile: "a/" + path,
		ToFile:   "b/" + path,
		Context:  3,
	}
	text, _ := difflib.GetUnifiedDiffString(d)
	if !truncate {
		return text
	}
	lines := strings.Split(text, "\n")
	if len(lines) <= maxDiffLinesPerFile {
		return text
	}
	trimmed := append(lines[:maxDiffLinesPerFile], "... (diff truncated)")
	return strings.Join(trimmed, "\n")
}
