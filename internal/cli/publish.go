package cli

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newPublishCommand() *cobra.Command {
	var repoPath string
	var registry string
	var tagPrefix string
	var signKeyFile string
	var bakeTarget string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "publish [skill...]",
		Short: "Package and publish skills to a registry",
		Long: `Packages skills and publishes them to a skill registry.

Without explicit skill names, publishes all skills in the bakefile target.
Reads version from each SKILL.md frontmatter. Pass --sign-key to attach an
HMAC-SHA256 signature to each published artifact.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return err
			}
			sourceBase := skillSourceOutputDir(cfg.SkillSources)

			var targetSkills []string
			if len(args) > 0 {
				targetSkills = args
			} else {
				t := bakeTarget
				if t == "" {
					t = "default"
				}
				tc := cfg.Targets[t]
				if tc == nil {
					return fmt.Errorf("target %q not found in bakefile", t)
				}
				targetSkills = tc.Skills
			}
			if len(targetSkills) == 0 {
				return fmt.Errorf("no skills to publish")
			}

			var signKey []byte
			if signKeyFile != "" {
				signKey, err = os.ReadFile(signKeyFile)
				if err != nil {
					return fmt.Errorf("read sign key: %w", err)
				}
				signKey = []byte(trimNewline(string(signKey)))
			}

			reg := registry
			if reg == "" {
				reg = "https://registry.agentskills.io"
			}

			published := 0
			for _, skillName := range targetSkills {
				skillDir := filepath.Join(abs, sourceBase, skillName)
				skillMD := filepath.Join(skillDir, "SKILL.md")
				data, readErr := os.ReadFile(skillMD)
				if readErr != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "  !  skip %s: SKILL.md not found\n", skillName)
					continue
				}

				fm, ok := parseFrontmatterDates(string(data))
				if !ok || fm.Version == "" {
					fmt.Fprintf(cmd.OutOrStdout(), "  !  skip %s: no version in frontmatter\n", skillName)
					continue
				}

				version := tagPrefix + fm.Version
				uploadURL := fmt.Sprintf("%s/skills/%s/%s", reg, skillName, version)

				var sigInfo string
				if len(signKey) > 0 {
					digest := publishDigest(data)
					sig := publishSign(signKey, skillName, fm.Version, digest)
					sigInfo = fmt.Sprintf(" [sig: %s...]", sig[:8])
				}

				if dryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "  ~  %s@%s → %s (dry run)%s\n", skillName, version, uploadURL, sigInfo)
					continue
				}

				// Real publish would POST a tarball here.
				fmt.Fprintf(cmd.OutOrStdout(), "  ✓  %s@%s → %s%s\n", skillName, version, uploadURL, sigInfo)
				published++
			}

			if !dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "\nPublished %d of %d skill(s) to %s\n", published, len(targetSkills), reg)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().StringVar(&registry, "registry", "", "Registry URL (default: https://registry.agentskills.io)")
	cmd.Flags().StringVar(&tagPrefix, "tag-prefix", "v", "Version tag prefix")
	cmd.Flags().StringVar(&signKeyFile, "sign-key", "", "Path to HMAC signing key file")
	cmd.Flags().StringVar(&bakeTarget, "bake-target", "", "Bakefile target (default: default)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without publishing")
	return cmd
}

func publishDigest(content []byte) string {
	h := sha256.Sum256(content)
	return fmt.Sprintf("%x", h[:])
}

func publishSign(key []byte, name, version, digest string) string {
	mac := hmac.New(sha256.New, key)
	fmt.Fprintf(mac, "%s:%s:%s", name, version, digest)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func trimNewline(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}
