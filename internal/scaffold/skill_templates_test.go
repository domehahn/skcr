package scaffold

import (
	"strings"
	"testing"
)

// requiredSections lists the markdown headings every registered SDLC skill must contain.
var requiredSections = []string{
	"## Skill-Specific Operating Model",
	"## Skill-Specific Checklist",
	"## Decision Rules",
	"## DevSecOps Guardrails",
	"## Output Requirements",
	"## Acceptance Criteria",
	"## Anti-Patterns",
	"## Changelog",
}

// requiredFrontmatterFields lists YAML keys that must appear in every rendered SKILL.md.
var requiredFrontmatterFields = []string{
	"version:",
	"since:",
	"last_modified:",
	"authors:",
	"stability:",
	"min_platform_version:",
	"changelog:",
}

func TestAllSDLCSkillsAreRegistered(t *testing.T) {
	for _, name := range SDLCSkillNames {
		if _, ok := skillBodies[name]; !ok {
			t.Errorf("SDLC skill %q is listed in SDLCSkillNames but has no registered template body", name)
		}
	}
	if len(SDLCSkillNames) != 18 {
		t.Errorf("expected 18 SDLC skills, got %d", len(SDLCSkillNames))
	}
}

func TestEveryRegisteredSkillRendersWithoutError(t *testing.T) {
	platforms := []string{"codex", "claude-code", "gitlab-duo"}
	for name := range skillBodies {
		t.Run(name, func(t *testing.T) {
			data := skillTemplateData{
				Name:         name,
				Title:        skillTitle(name),
				Description:  "Test description for " + name,
				Version:      "1.0.0",
				Since:        "2025-01-01",
				LastModified: "2026-06-10",
				Owner:        "platform-engineering",
				Stability:    "stable",
				License:      "MIT",
				Platforms:    platforms,
			}
			rendered, err := renderSkillTemplate(name, data)
			if err != nil {
				t.Fatalf("renderSkillTemplate(%q) returned error: %v", name, err)
			}
			if rendered == "" {
				t.Fatalf("renderSkillTemplate(%q) returned empty string", name)
			}
		})
	}
}

func TestEveryRegisteredSkillContainsRequiredSections(t *testing.T) {
	platforms := []string{"codex", "claude-code"}
	for name := range skillBodies {
		t.Run(name, func(t *testing.T) {
			data := skillTemplateData{
				Name:         name,
				Title:        skillTitle(name),
				Description:  "Test description",
				Version:      "1.0.0",
				Since:        "2025-01-01",
				LastModified: "2026-06-10",
				Owner:        "platform-engineering",
				Stability:    "experimental",
				License:      "MIT",
				Platforms:    platforms,
			}
			rendered, err := renderSkillTemplate(name, data)
			if err != nil {
				t.Fatalf("render error: %v", err)
			}
			for _, section := range requiredSections {
				if !strings.Contains(rendered, section) {
					t.Errorf("skill %q is missing required section %q", name, section)
				}
			}
		})
	}
}

func TestEveryRegisteredSkillContainsRequiredFrontmatterFields(t *testing.T) {
	platforms := []string{"codex", "claude-code"}
	for name := range skillBodies {
		t.Run(name, func(t *testing.T) {
			data := skillTemplateData{
				Name:         name,
				Title:        skillTitle(name),
				Description:  "Test description",
				Version:      "0.9.0",
				Since:        "2025-06-01",
				LastModified: "2026-06-10",
				Owner:        "platform-engineering",
				Stability:    "experimental",
				License:      "MIT",
				Platforms:    platforms,
			}
			rendered, err := renderSkillTemplate(name, data)
			if err != nil {
				t.Fatalf("render error: %v", err)
			}
			for _, field := range requiredFrontmatterFields {
				if !strings.Contains(rendered, field) {
					t.Errorf("skill %q is missing required frontmatter field %q", name, field)
				}
			}
		})
	}
}

func TestFrontmatterContainsTemplateVariables(t *testing.T) {
	platforms := []string{"codex", "gitlab-duo", "cursor"}
	data := skillTemplateData{
		Name:         "requirements-analyst",
		Title:        "Requirements Analyst",
		Description:  "Analyze requirements for testing",
		Version:      "2.3.4",
		Since:        "2024-03-15",
		LastModified: "2026-06-10",
		Owner:        "security-team",
		Stability:    "stable",
		License:      "MIT",
		Platforms:    platforms,
	}
	rendered, err := renderSkillTemplate("requirements-analyst", data)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`name: requirements-analyst`,
		`version: "2.3.4"`,
		`since: "2024-03-15"`,
		`last_modified: "2026-06-10"`,
		`- security-team`,
		`stability: stable`,
		`codex: "unknown"`,
		`gitlab-duo: "unknown"`,
		`cursor: "unknown"`,
		`# Requirements Analyst`,
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("expected rendered output to contain %q\nfull output:\n%s", want, rendered[:min(500, len(rendered))])
		}
	}
}

func TestSkillsAreNotGenericCopies(t *testing.T) {
	platforms := []string{"codex"}
	render := func(name string) string {
		data := skillTemplateData{
			Name: name, Title: skillTitle(name),
			Description: "x", Version: "1.0.0", Since: "2025-01-01",
			LastModified: "2026-06-10", Owner: "platform-engineering",
			Stability: "experimental", License: "MIT", Platforms: platforms,
		}
		out, _ := renderSkillTemplate(name, data)
		return out
	}

	// Strip frontmatter and title — compare only the body sections.
	body := func(s string) string {
		idx := strings.Index(s, "\n# ")
		if idx < 0 {
			return s
		}
		return s[idx:]
	}

	names := SDLCSkillNames
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			a, b := names[i], names[j]
			bodyA := body(render(a))
			bodyB := body(render(b))
			if bodyA == bodyB {
				t.Errorf("skills %q and %q produce identical bodies — templates must be skill-specific", a, b)
			}
		}
	}
}

func TestEachSkillHasUniqueChecklistItem(t *testing.T) {
	platforms := []string{"codex"}
	// Collect all checklist lines across skills
	allLines := map[string][]string{} // skill -> checklist lines
	for _, name := range SDLCSkillNames {
		data := skillTemplateData{
			Name: name, Title: skillTitle(name),
			Description: "x", Version: "1.0.0", Since: "2025-01-01",
			LastModified: "2026-06-10", Owner: "platform-engineering",
			Stability: "experimental", License: "MIT", Platforms: platforms,
		}
		rendered, _ := renderSkillTemplate(name, data)
		for _, line := range strings.Split(rendered, "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "- [ ]") {
				allLines[name] = append(allLines[name], strings.TrimSpace(line))
			}
		}
	}

	// Every skill must have at least 10 checklist items.
	for _, name := range SDLCSkillNames {
		if len(allLines[name]) < 10 {
			t.Errorf("skill %q has only %d checklist items, need at least 10", name, len(allLines[name]))
		}
	}

	// Each skill must have at least one checklist item not shared by any other skill.
	for _, name := range SDLCSkillNames {
		hasUnique := false
		for _, line := range allLines[name] {
			count := 0
			for _, other := range SDLCSkillNames {
				for _, otherLine := range allLines[other] {
					if line == otherLine {
						count++
					}
				}
			}
			if count == 1 {
				hasUnique = true
				break
			}
		}
		if !hasUnique {
			t.Errorf("skill %q has no unique checklist item — all its checklist items are shared with other skills", name)
		}
	}
}

func TestUniversalSkillCreatorForbidsGenericCopyPaste(t *testing.T) {
	// universal-skill-creator is handled by the SKILL.md in .agents/skills/.
	// For the scaffold template registry, verify requirements-analyst (or any
	// registered skill) does not contain the generic placeholder text.
	for _, name := range SDLCSkillNames {
		data := skillTemplateData{
			Name: name, Title: skillTitle(name),
			Description: "x", Version: "1.0.0", Since: "2025-01-01",
			LastModified: "2026-06-10", Owner: "platform-engineering",
			Stability: "experimental", License: "MIT",
			Platforms: []string{"codex"},
		}
		rendered, _ := renderSkillTemplate(name, data)
		if strings.Contains(rendered, "Describe what this skill helps an agent do") {
			t.Errorf("skill %q contains generic placeholder text", name)
		}
		if strings.Contains(rendered, "TODO") {
			t.Errorf("skill %q contains TODO placeholder", name)
		}
		if strings.Contains(rendered, "Add checks here") {
			t.Errorf("skill %q contains generic placeholder checklist text", name)
		}
	}
}

func TestGenericFallbackRendersForUnknownSkill(t *testing.T) {
	// renderSkillTemplate returns ("", nil) for unknown skills — skillMarkdown falls back.
	rendered, err := renderSkillTemplate("some-future-skill", skillTemplateData{
		Name: "some-future-skill", Title: "Some Future Skill",
		Platforms: []string{"codex"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rendered != "" {
		t.Errorf("expected empty string for unknown skill, got non-empty output")
	}
}

func TestSkillTitleConversion(t *testing.T) {
	cases := []struct{ in, want string }{
		{"requirements-analyst", "Requirements Analyst"},
		{"ci-cd-security-reviewer", "Ci Cd Security Reviewer"},
		{"iac-security-reviewer", "Iac Security Reviewer"},
		{"safe-implementer", "Safe Implementer"},
	}
	for _, c := range cases {
		got := skillTitle(c.in)
		if got != c.want {
			t.Errorf("skillTitle(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPlanSkillUsesRegisteredTemplate(t *testing.T) {
	files, err := PlanSkill(SkillOptions{
		Name:      "requirements-analyst",
		OutputDir: t.TempDir(),
		Version:   "1.0.0",
		Platforms: []string{"codex", "claude-code"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var skillMD string
	for _, f := range files {
		if strings.HasSuffix(f.Path, "SKILL.md") {
			skillMD = f.Content
		}
	}
	if skillMD == "" {
		t.Fatal("SKILL.md not found in planned files")
	}
	for _, section := range requiredSections {
		if !strings.Contains(skillMD, section) {
			t.Errorf("PlanSkill(requirements-analyst) SKILL.md missing section %q", section)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
