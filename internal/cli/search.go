package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newSearchCommand() *cobra.Command {
	var repoPath string
	var stability string
	var author string
	var query string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search skills by metadata fields",
		Long: `Filters the skill list by frontmatter metadata.
Multiple flags are ANDed: --stability experimental --author alice
shows only experimental skills authored by alice.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}
			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}
			sourceBase := skillSourceOutputDir(cfg.SkillSources)

			skillTargets := map[string][]string{}
			for tn, t := range cfg.Targets {
				for _, s := range t.Skills {
					skillTargets[s] = append(skillTargets[s], tn)
				}
			}

			var names []string
			for name := range skillTargets {
				names = append(names, name)
			}
			sort.Strings(names)

			type matchEntry struct {
				Name        string   `json:"name"`
				Version     string   `json:"version,omitempty"`
				Description string   `json:"description,omitempty"`
				Stability   string   `json:"stability,omitempty"`
				Authors     []string `json:"authors,omitempty"`
				Targets     []string `json:"targets,omitempty"`
			}

			var matches []matchEntry
			for _, name := range names {
				skillMD := filepath.Join(absTarget, sourceBase, name, "SKILL.md")
				data, readErr := os.ReadFile(skillMD)
				if readErr != nil {
					continue
				}

				var fm map[string]any
				content := string(data)
				if strings.HasPrefix(content, "---\n") {
					end := strings.Index(content[4:], "\n---")
					if end >= 0 {
						_ = yaml.Unmarshal([]byte(content[4:4+end]), &fm)
					}
				}

				desc, _ := fm["description"].(string)
				stab, _ := fm["stability"].(string)
				ver, _ := fm["version"].(string)

				var authors []string
				switch v := fm["authors"].(type) {
				case []any:
					for _, a := range v {
						if s, ok := a.(string); ok {
							authors = append(authors, s)
						}
					}
				case string:
					if v != "" {
						authors = []string{v}
					}
				}

				// Apply filters (all must match).
				if stability != "" && !strings.EqualFold(stab, stability) {
					continue
				}
				if author != "" {
					found := false
					for _, a := range authors {
						if strings.Contains(strings.ToLower(a), strings.ToLower(author)) {
							found = true
							break
						}
					}
					if !found {
						continue
					}
				}
				if query != "" {
					q := strings.ToLower(query)
					if !strings.Contains(strings.ToLower(name+" "+desc), q) {
						continue
					}
				}

				targets := append([]string(nil), skillTargets[name]...)
				sort.Strings(targets)

				matches = append(matches, matchEntry{
					Name:        name,
					Version:     ver,
					Description: desc,
					Stability:   stab,
					Authors:     authors,
					Targets:     targets,
				})
			}

			if len(matches) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No skills matched.")
				return nil
			}

			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(matches)
			}

			for _, m := range matches {
				line := fmt.Sprintf("%-40s  %-12s  %-14s", m.Name, m.Version, m.Stability)
				if m.Description != "" {
					line += "  " + m.Description
				}
				fmt.Fprintln(cmd.OutOrStdout(), strings.TrimRight(line, " "))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n%d match(es).\n", len(matches))
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().StringVar(&stability, "stability", "", "Filter by stability (stable, experimental, deprecated)")
	cmd.Flags().StringVar(&author, "author", "", "Filter by author name (substring match)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Full-text search across name and description")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	return cmd
}
