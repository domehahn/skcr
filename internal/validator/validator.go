package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/agentic-template-kit/skcr/internal/models"
	"gopkg.in/yaml.v3"
)

func ValidateProject(target string) ([]string, error) {
	errors := []string{}

	bakePath := filepath.Join(target, "agentic.bake.yaml")
	if _, err := os.Stat(bakePath); err != nil {
		if os.IsNotExist(err) {
			return []string{"Missing agentic.bake.yaml"}, nil
		}
		return nil, err
	}

	payload, err := os.ReadFile(bakePath)
	if err != nil {
		return nil, err
	}
	raw := map[string]any{}
	if err := yaml.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}

	targets, _ := raw["targets"].(map[string]any)
	if len(targets) == 0 {
		errors = append(errors, "No targets configured in agentic.bake.yaml")
	}

	for name, cfgRaw := range targets {
		cfg, ok := cfgRaw.(map[string]any)
		if !ok {
			continue
		}
		platforms, _ := cfg["platforms"].([]any)
		for _, p := range platforms {
			platform, _ := p.(string)
			if _, ok := models.SupportedPlatforms[platform]; !ok {
				errors = append(errors, fmt.Sprintf("Target %s: unsupported platform %s", name, platform))
			}
		}
	}

	for _, baseDir := range []string{"skills", ".agents/skills", ".claude/skills", ".agentic/skills"} {
		skillsDir := filepath.Join(target, baseDir)
		if entries, err := os.ReadDir(skillsDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				skillFile := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
				text, err := os.ReadFile(skillFile)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Skill missing SKILL.md: %s", filepath.Dir(skillFile)))
					continue
				}
				if errMsg := validateSkillMetadata(string(text)); errMsg != "" {
					errors = append(errors, fmt.Sprintf("%s: %s", errMsg, skillFile))
				}
			}
		}
	}

	gitlabDuo := filepath.Join(target, ".gitlab", "duo")
	if stat, err := os.Stat(gitlabDuo); err == nil && stat.IsDir() {
		chatRules := filepath.Join(gitlabDuo, "chat-rules.md")
		if _, err := os.Stat(chatRules); err != nil {
			errors = append(errors, "GitLab Duo output missing .gitlab/duo/chat-rules.md")
		}

		flowDir := filepath.Join(gitlabDuo, "flows")
		if entries, err := os.ReadDir(flowDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
					continue
				}
				flowPath := filepath.Join(flowDir, entry.Name())
				text, err := os.ReadFile(flowPath)
				if err != nil {
					continue
				}
				s := string(text)
				for _, forbidden := range []string{"name:", "description:", "product_group:"} {
					if len(s) >= len(forbidden) && s[:len(forbidden)] == forbidden {
						errors = append(errors, fmt.Sprintf("GitLab custom flow contains forbidden top-level field %s: %s", forbidden, flowPath))
					}
				}
				if !containsAll(s, []string{"workspace_agent_skills"}) {
					errors = append(errors, fmt.Sprintf("Flow does not pass workspace_agent_skills: %s", flowPath))
				}
				if !containsAll(s, []string{"user_rule"}) {
					errors = append(errors, fmt.Sprintf("Flow does not pass user_rule: %s", flowPath))
				}
			}
		}
	}

	return errors, nil
}

func containsAll(s string, terms []string) bool {
	for _, t := range terms {
		if !strings.Contains(s, t) {
			return false
		}
	}
	return true
}

var (
	nameRegex        = regexp.MustCompile(`(?m)^name:\s*(.+?)\s*$`)
	descriptionRegex = regexp.MustCompile(`(?m)^description:\s*(.+?)\s*$`)
)

func validateSkillMetadata(content string) string {
	nameMatch := nameRegex.FindStringSubmatch(content)
	if len(nameMatch) < 2 || isEmptyMetadataValue(nameMatch[1]) {
		return "Skill metadata name is missing or empty"
	}

	descriptionMatch := descriptionRegex.FindStringSubmatch(content)
	if len(descriptionMatch) < 2 || isEmptyMetadataValue(descriptionMatch[1]) {
		return "Skill metadata description is missing or empty"
	}

	return ""
}

func isEmptyMetadataValue(value string) bool {
	v := strings.TrimSpace(value)
	v = strings.Trim(v, `"'`)
	return strings.TrimSpace(v) == ""
}
