package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newUpgradeSkillCommand() *cobra.Command {
	var target string
	var message string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "upgrade <skill> <version>",
		Short: "Bump a skill's version and record a changelog entry in SKILL.md",
		Long: `Sets version: in SKILL.md frontmatter to <version> and appends a new entry
to the changelog: list in the same frontmatter block.

This is the combined equivalent of running 'skcr pin' followed by editing the
changelog: block manually.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName, version := args[0], args[1]

			abs, err := filepath.Abs(target)
			if err != nil {
				return fmt.Errorf("resolve target: %w", err)
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return fmt.Errorf("load bakefile: %w", err)
			}

			sourceBase := skillSourceOutputDir(cfg.SkillSources)
			skillMD := filepath.Join(abs, sourceBase, skillName, "SKILL.md")

			raw, err := cliReadFile(skillMD)
			if err != nil {
				return fmt.Errorf("read SKILL.md: %w", err)
			}
			content := string(raw)

			// Pin the version.
			updated, _ := setFrontmatterField(content, "version", version)

			// Append to changelog: block in frontmatter.
			msg := message
			if msg == "" {
				msg = "Upgrade to " + version
			}
			updated, err = appendFrontmatterChangelog(updated, version, msg)
			if err != nil {
				return fmt.Errorf("update changelog: %w", err)
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Would upgrade %s to %s\n", skillName, version)
				return nil
			}

			if err := cliWriteFile(skillMD, []byte(updated), 0o644); err != nil {
				return fmt.Errorf("write SKILL.md: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Upgraded %s → %s\n", skillName, version)
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().StringVar(&message, "message", "", "Changelog entry message (default: 'Upgrade to <version>')")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	return cmd
}

// appendFrontmatterChangelog appends a new entry to the changelog: list in the
// SKILL.md frontmatter. If changelog: does not exist it is created.
func appendFrontmatterChangelog(content, version, message string) (string, error) {
	if !strings.HasPrefix(content, "---\n") {
		return content, fmt.Errorf("no YAML frontmatter found")
	}
	end := strings.Index(content[4:], "\n---")
	if end < 0 {
		return content, fmt.Errorf("frontmatter closing delimiter not found")
	}
	fmRaw := content[4 : 4+end]

	var fm yaml.Node
	if err := yaml.Unmarshal([]byte(fmRaw), &fm); err != nil {
		return content, err
	}

	// Parse into a generic map to preserve existing fields.
	var fmMap map[string]any
	if err := yaml.Unmarshal([]byte(fmRaw), &fmMap); err != nil {
		return content, err
	}

	today := time.Now().UTC().Format("2006-01-02")
	entry := map[string]any{
		"version": version,
		"date":    today,
		"change":  message,
	}

	existing, _ := fmMap["changelog"].([]any)
	fmMap["changelog"] = append(existing, entry)

	newFM, err := yaml.Marshal(fmMap)
	if err != nil {
		return content, err
	}

	body := content[4+end+4:]
	return "---\n" + string(newFM) + "---" + body, nil
}
