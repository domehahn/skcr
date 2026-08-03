package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/domehahn/skcr/internal/skillmeta"
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
	cmd.AddCommand(newMigrateSkillCommand())
	return cmd
}

func newMigrateSkillCommand() *cobra.Command {
	var target string
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "skill <name>",
		Short: "Explicitly migrate a legacy skill to split Descriptor/Contract/Eval artifacts",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := filepath.Abs(target)
			if err != nil {
				return err
			}
			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}
			skillDir := filepath.Join(absTarget, skillSourceOutputDir(cfg.SkillSources), args[0])
			if err := ensureMigrationRoot(absTarget, skillDir); err != nil {
				return err
			}
			legacyPath := skillmeta.DescriptorPath(skillDir)
			legacy, err := skillmeta.LoadDescriptor(legacyPath)
			if err != nil {
				return err
			}
			if legacy.SchemaVersion == skillmeta.DescriptorSchemaVersion && legacy.LegacyEmbeddedContract == nil && legacy.Contract != nil {
				if _, err := os.Stat(filepath.Join(skillDir, legacy.Contract.File)); err == nil {
					if filepath.Base(legacyPath) == skillmeta.DescriptorFilename {
						fmt.Fprintln(cmd.OutOrStdout(), "No migration needed — split artifact already exists.")
						return nil
					}
				}
			}
			description := legacy.Description
			if strings.TrimSpace(description) == "" {
				description = "TODO: Describe what this skill helps an agent do."
			}
			descriptor := skillmeta.NewDescriptor(legacy.Name, legacy.Version, description, legacy.License, legacy.Owners, legacy.CompatibleWith)
			descriptor.Namespace = legacy.Namespace
			descriptor.Tags = legacy.Tags
			descriptor.Security = legacy.Security
			descriptor.Metadata = legacy.Metadata
			if legacy.Goal != nil && strings.TrimSpace(legacy.Goal.Objective) != "" {
				descriptor.Goal = legacy.Goal
			}
			contract := skillmeta.NewContractV2() // explicit default deny; never inferred from prose
			eval := skillmeta.NewBaselineEvalV2()
			mcp := skillmeta.NewMCPDocument()
			a2a := skillmeta.NewA2ADocument(descriptor.Name)
			dependencies := skillmeta.NewDependenciesDocument()
			assurance := skillmeta.NewAssuranceDocument()
			if legacy.Security != nil {
				fmt.Fprintln(cmd.OutOrStdout(), "Warning: preserved deprecated skill.yaml security hints for compatibility; they do not grant permissions and contract.yaml remains authoritative.")
			}
			files := []struct {
				path      string
				value     any
				overwrite bool
			}{
				{filepath.Join(skillDir, skillmeta.DescriptorFilename), descriptor, true},
				{filepath.Join(skillDir, "contract.yaml"), contract, false},
				{filepath.Join(skillDir, "evals", "baseline.yaml"), eval, false},
				{filepath.Join(skillDir, "integrations", "mcp.yaml"), mcp, false},
				{filepath.Join(skillDir, "integrations", "a2a.yaml"), a2a, false},
				{filepath.Join(skillDir, "dependencies.yaml"), dependencies, false},
				{filepath.Join(skillDir, "assurance.yaml"), assurance, false},
			}
			if dryRun {
				for _, file := range files {
					if _, err := os.Stat(file.path); err == nil && !file.overwrite {
						fmt.Fprintf(cmd.OutOrStdout(), "would preserve  %s\n", file.path)
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "would write  %s\n", file.path)
					}
				}
				if filepath.Base(legacyPath) == skillmeta.LegacyDescriptorFilename {
					fmt.Fprintf(cmd.OutOrStdout(), "would remove  %s after successful rename\n", legacyPath)
				}
				return nil
			}
			for _, file := range files {
				if _, err := os.Stat(file.path); err == nil && !file.overwrite {
					continue
				}
				data, err := yaml.Marshal(file.value)
				if err != nil {
					return err
				}
				if err := os.MkdirAll(filepath.Dir(file.path), 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(file.path, data, 0o644); err != nil {
					return err
				}
			}
			readme := filepath.Join(skillDir, "evals", "README.md")
			if _, err := os.Stat(readme); os.IsNotExist(err) {
				if err := os.WriteFile(readme, []byte("# Behavioral Evals\n\nDeclarative behavioral and adversarial scenarios for this skill.\n"), 0o644); err != nil {
					return err
				}
			}
			if filepath.Base(legacyPath) == skillmeta.LegacyDescriptorFilename {
				if err := os.Remove(legacyPath); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("remove migrated legacy descriptor %s: %w", legacyPath, err)
				}
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Migrated skill with a default-deny Contract; review TODO Goal and capability scopes explicitly.")
			return nil
		},
	}
	cmd.Flags().StringVarP(&target, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	return cmd
}

func ensureMigrationRoot(projectRoot, skillDir string) error {
	resolvedRoot, err := filepath.EvalSymlinks(projectRoot)
	if err != nil {
		return err
	}
	resolvedSkill, err := filepath.EvalSymlinks(skillDir)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(resolvedRoot, resolvedSkill)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("skill directory resolves outside project root")
	}
	for _, path := range []string{filepath.Join(skillDir, "contract.yaml"), filepath.Join(skillDir, "evals"), filepath.Join(skillDir, "integrations"), filepath.Join(skillDir, "dependencies.yaml"), filepath.Join(skillDir, "assurance.yaml")} {
		if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("migration refuses symlinked artifact path: %s", path)
		}
	}
	return nil
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
