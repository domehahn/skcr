package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newShowCommand() *cobra.Command {
	var repoPath string
	var previewLines int
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "show <skill>",
		Short: "Display a skill's frontmatter and body preview",
		Long: `Reads the SKILL.md of the named skill and pretty-prints its frontmatter fields
along with the first lines of the body content.

Use --json to get all fields as a JSON object.
Use --preview-lines to control how many body lines are shown (default: 8).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName := args[0]

			abs, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return err
			}
			sourceBase := skillSourceOutputDir(cfg.SkillSources)
			skillMDPath := filepath.Join(abs, sourceBase, skillName, "SKILL.md")

			data, err := os.ReadFile(skillMDPath)
			if err != nil {
				return fmt.Errorf("SKILL.md not found for %q: %w", skillName, err)
			}

			content := string(data)
			fm, body := splitFrontmatterBody(content)

			if jsonOut {
				var fmMap map[string]any
				_ = yaml.Unmarshal([]byte(fm), &fmMap)
				out := map[string]any{
					"name":     skillName,
					"path":     skillMDPath,
					"metadata": fmMap,
					"preview":  bodyPreview(body, previewLines),
				}
				enc, _ := json.MarshalIndent(out, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(enc))
				return nil
			}

			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Skill: %s\n", skillName)
			fmt.Fprintf(w, "Path:  %s\n\n", skillMDPath)

			// Print frontmatter fields.
			var fmMap map[string]any
			if err := yaml.Unmarshal([]byte(fm), &fmMap); err == nil {
				keys := []string{"name", "description", "version", "stability", "license",
					"since", "last_modified", "compatible_with", "tags"}
				for _, k := range keys {
					v, ok := fmMap[k]
					if !ok || v == nil {
						continue
					}
					switch val := v.(type) {
					case []any:
						parts := make([]string, len(val))
						for i, p := range val {
							parts[i] = fmt.Sprintf("%v", p)
						}
						fmt.Fprintf(w, "  %-18s %s\n", k+":", strings.Join(parts, ", "))
					default:
						fmt.Fprintf(w, "  %-18s %v\n", k+":", val)
					}
				}
			}

			if previewLines > 0 && strings.TrimSpace(body) != "" {
				fmt.Fprintf(w, "\n--- body preview (%d lines) ---\n", previewLines)
				fmt.Fprintln(w, bodyPreview(body, previewLines))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().IntVar(&previewLines, "preview-lines", 8, "Number of body lines to preview (0 = none)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

// splitFrontmatterBody separates the YAML frontmatter from the body of a SKILL.md.
func splitFrontmatterBody(content string) (fm, body string) {
	if !strings.HasPrefix(content, "---\n") {
		return "", content
	}
	end := strings.Index(content[4:], "\n---")
	if end < 0 {
		return "", content
	}
	fm = content[4 : 4+end]
	body = content[4+end+4:]
	return
}

// bodyPreview returns the first n non-empty lines of body text.
func bodyPreview(body string, n int) string {
	lines := strings.Split(strings.TrimSpace(body), "\n")
	if n > len(lines) {
		n = len(lines)
	}
	return strings.Join(lines[:n], "\n")
}
