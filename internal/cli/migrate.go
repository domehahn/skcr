package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newMigrateCommand() *cobra.Command {
	var repoPath string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate agentic.bake.yaml to the current format",
		Long: `Applies format migrations to agentic.bake.yaml in place.
Currently handles:
  - skill_sources.defaults.version  →  skill_sources.defaults.initial_version
  - skill_sources.skills[n].version →  skill_sources.skills[n].initial_version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}
			bakePath := filepath.Join(absTarget, "agentic.bake.yaml")
			raw, err := os.ReadFile(bakePath)
			if err != nil {
				return fmt.Errorf("cannot read %s: %w", bakePath, err)
			}

			var doc yaml.Node
			if err := yaml.Unmarshal(raw, &doc); err != nil {
				return fmt.Errorf("cannot parse %s: %w", bakePath, err)
			}

			changes := migrateYAMLNode(&doc)

			if len(changes) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No migrations needed — agentic.bake.yaml is already up to date.")
				return nil
			}

			for _, c := range changes {
				fmt.Fprintf(cmd.OutOrStdout(), "  rename  %s  →  %s\n", c.from, c.to)
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "\nDry run: %d change(s) would be applied. Re-run without --dry-run to write.\n", len(changes))
				return nil
			}

			out, err := yaml.Marshal(doc.Content[0])
			if err != nil {
				return fmt.Errorf("cannot marshal updated bakefile: %w", err)
			}
			if err := os.WriteFile(bakePath, out, 0o644); err != nil {
				return fmt.Errorf("cannot write %s: %w", bakePath, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nApplied %d migration(s) to agentic.bake.yaml.\n", len(changes))
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview migrations without writing")
	return cmd
}

type yamlRename struct{ from, to string }

// migrateYAMLNode walks the YAML document and renames deprecated keys in place.
func migrateYAMLNode(doc *yaml.Node) []yamlRename {
	var changes []yamlRename
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		changes = walkMigrate(doc.Content[0], []string{}, changes)
	}
	return changes
}

// renamedKeys maps deprecated key paths to their replacement key names.
// Key: dot-separated path to the key, Value: new key name.
var renamedKeys = map[string]string{
	"skill_sources.defaults.version": "initial_version",
}

// renamedKeysByParent matches keys by parent path prefix for array items.
var renamedKeysByParent = map[string]map[string]string{
	"skill_sources.skills": {"version": "initial_version"},
}

func walkMigrate(node *yaml.Node, path []string, changes []yamlRename) []yamlRename {
	if node.Kind != yaml.MappingNode {
		return changes
	}
	parentPath := strings.Join(path, ".")
	for i := 0; i+1 < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valNode := node.Content[i+1]
		childPath := append(append([]string(nil), path...), keyNode.Value)
		fullPath := strings.Join(childPath, ".")

		// Direct key rename.
		if newKey, ok := renamedKeys[fullPath]; ok {
			changes = append(changes, yamlRename{from: fullPath, to: strings.Join(append(path, newKey), ".")})
			keyNode.Value = newKey
		}

		// Recurse into mapping values.
		if valNode.Kind == yaml.MappingNode {
			changes = walkMigrate(valNode, childPath, changes)
		}

		// Recurse into sequence items (for array-element key renames).
		if valNode.Kind == yaml.SequenceNode {
			if renames, ok := renamedKeysByParent[fullPath]; ok {
				for _, item := range valNode.Content {
					if item.Kind != yaml.MappingNode {
						continue
					}
					for j := 0; j+1 < len(item.Content); j += 2 {
						k := item.Content[j]
						if newKey, ok := renames[k.Value]; ok {
							fromPath := fullPath + "[*]." + k.Value
							toPath := fullPath + "[*]." + newKey
							changes = append(changes, yamlRename{from: fromPath, to: toPath})
							k.Value = newKey
						}
					}
				}
			}
			for _, item := range valNode.Content {
				changes = walkMigrate(item, childPath, changes)
			}
		}

		_ = parentPath
	}
	return changes
}
