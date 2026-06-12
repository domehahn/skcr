package scaffold

import (
	"regexp"
	"strings"
	"testing"

	"github.com/domehahn/skcr/internal/platforms"
)

const testSince = "2026-06-10"
const testLastModified = "2026-06-12"
const testVersion = "1.0.0"

var dateRE = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

var expectedSDLCSkills = []string{
	"requirements-analyst",
	"cost-based-planner",
	"architecture-reviewer",
	"threat-modeler",
	"safe-implementer",
	"test-strategy-engineer",
	"verification-reviewer",
	"security-reviewer",
	"secrets-reviewer",
	"dependency-supply-chain-reviewer",
	"ci-cd-reviewer",
	"iac-gitops-reviewer",
	"compliance-governance-reviewer",
	"release-readiness-reviewer",
	"observability-reviewer",
	"incident-postmortem-assistant",
	"documentation-maintainer",
	"universal-skill-creator",
}

var requiredSections = []string{
	"## Purpose",
	"## When to use",
	"## Operating model",
	"## Spec-Driven Change Context",
	"## Skill-Specific Review Scope",
	"## Skill-Specific Checklist",
	"## Decision Rules",
	"## Finding Categories",
	"## Severity Guidance",
	"## DevSecOps Guardrails",
	"## Output Requirements",
	"## Acceptance Criteria",
	"## Anti-Patterns",
	"## Changelog",
}

var requiredFrontmatterFields = []string{
	"version:",
	"since:",
	"last_modified:",
	"authors:",
	"stability:",
	"min_platform_version:",
	"changelog:",
}

func TestAllConfiguredSDLCSkillsAreRegistered(t *testing.T) {
	if len(SDLCSkillNames) != len(expectedSDLCSkills) {
		t.Fatalf("expected %d SDLC skills, got %d: %#v", len(expectedSDLCSkills), len(SDLCSkillNames), SDLCSkillNames)
	}
	for i, want := range expectedSDLCSkills {
		if SDLCSkillNames[i] != want {
			t.Fatalf("SDLCSkillNames[%d] = %q, want %q", i, SDLCSkillNames[i], want)
		}
		if _, ok := skillBodies[want]; !ok {
			t.Fatalf("skill %q is listed but has no registered template body", want)
		}
	}
}

func TestEveryRegisteredSkillRendersProductionReadySkillMD(t *testing.T) {
	for _, name := range expectedSDLCSkills {
		t.Run(name, func(t *testing.T) {
			rendered := renderForTest(t, name)
			if !strings.HasPrefix(rendered, "---\n") {
				t.Fatalf("rendered skill %q does not start with YAML frontmatter", name)
			}
			for _, field := range requiredFrontmatterFields {
				if !strings.Contains(rendered, field) {
					t.Errorf("skill %q missing frontmatter field %q", name, field)
				}
			}
			for _, platform := range platforms.AllMinVersions() {
				want := platform.Name + `: "` + platform.MinVersion + `"`
				if !strings.Contains(rendered, want) {
					t.Errorf("skill %q missing min_platform_version entry %q", name, want)
				}
				if platform.Status == "verified" && platform.MinVersion == "unknown" {
					t.Errorf("verified platform %q must not use unknown", platform.Name)
				}
			}
			for _, section := range requiredSections {
				if !strings.Contains(rendered, section) {
					t.Errorf("skill %q missing required section %q", name, section)
				}
			}
			date := extractFrontmatterDate(rendered)
			if date == "" {
				t.Errorf("skill %q frontmatter changelog missing date field", name)
			} else if !dateRE.MatchString(date) {
				t.Errorf("skill %q frontmatter changelog date %q is not YYYY-MM-DD", name, date)
			} else if !strings.Contains(rendered, "### "+testVersion+" - "+date) {
				t.Errorf("skill %q body changelog heading does not match frontmatter date %q", name, date)
			}
			if !strings.Contains(rendered, `change: "Initial generated production-ready SDLC / DevSecOps skill"`) ||
				!strings.Contains(rendered, "- Initial generated production-ready SDLC / DevSecOps skill.") {
				t.Errorf("skill %q frontmatter and body changelog messages are not synchronized", name)
			}
		})
	}
}

func TestEverySkillMeetsMinimumContentCounts(t *testing.T) {
	for _, name := range expectedSDLCSkills {
		t.Run(name, func(t *testing.T) {
			rendered := renderForTest(t, name)
			checks := map[string]int{
				"## Skill-Specific Checklist": 10,
				"## Decision Rules":           5,
				"## Finding Categories":       5,
				"## Severity Guidance":        4,
				"## Output Requirements":      5,
				"## Acceptance Criteria":      5,
				"## Anti-Patterns":            5,
			}
			for heading, want := range checks {
				got := countBullets(section(rendered, heading))
				if got < want {
					t.Errorf("skill %q section %q has %d bullets, want at least %d", name, heading, got, want)
				}
			}
		})
	}
}

func TestSkillBodiesAreNotNearlyIdentical(t *testing.T) {
	for i := 0; i < len(expectedSDLCSkills); i++ {
		for j := i + 1; j < len(expectedSDLCSkills); j++ {
			a := normalizedBody(renderForTest(t, expectedSDLCSkills[i]))
			b := normalizedBody(renderForTest(t, expectedSDLCSkills[j]))
			if a == b {
				t.Fatalf("skills %q and %q produce identical normalized bodies", expectedSDLCSkills[i], expectedSDLCSkills[j])
			}
			if similarity(a, b) > 0.92 {
				t.Fatalf("skills %q and %q are too similar", expectedSDLCSkills[i], expectedSDLCSkills[j])
			}
		}
	}
}

func TestRequiredDomainTerms(t *testing.T) {
	cases := map[string][]string{
		"requirements-analyst":             {"functional requirements", "non-functional requirements", "acceptance criteria", "traceability", "stakeholders"},
		"cost-based-planner":               {"cost drivers", "budget", "scope", "trade-offs", "estimate"},
		"architecture-reviewer":            {"circular dependencies", "coupling", "module boundaries", "data ownership", "ADR"},
		"threat-modeler":                   {"assets", "trust boundaries", "entry points", "STRIDE", "abuse cases"},
		"safe-implementer":                 {"minimal", "rollback", "input validation", "tests", "broad refactoring"},
		"test-strategy-engineer":           {"contract tests", "negative tests", "fixtures", "CI gates", "flaky"},
		"verification-reviewer":            {"acceptance criteria", "validation evidence", "generated outputs", "residual risk", "conditional pass"},
		"security-reviewer":                {"AuthZ bypass", "injection", "SSRF", "path traversal", "unsafe logging", "insecure defaults"},
		"secrets-reviewer":                 {"rotation", "revocation", "CI/CD secret", "redaction", "false positive"},
		"dependency-supply-chain-reviewer": {"SBOM", "provenance", "lockfile", "reachability", "license"},
		"ci-cd-reviewer":                   {"token permissions", "cache poisoning", "artifact integrity", "runner trust", "deployment gates"},
		"iac-gitops-reviewer":              {"Terraform", "Kubernetes", "IAM/RBAC", "drift", "rollback"},
		"compliance-governance-reviewer":   {"segregation of duties", "audit evidence", "risk acceptance", "policy exception", "retention"},
		"release-readiness-reviewer":       {"go/no-go", "rollback", "monitoring", "known issues", "feature flag"},
		"observability-reviewer":           {"SLO", "SLI", "correlation IDs", "cardinality", "alert routing"},
		"incident-postmortem-assistant":    {"timeline", "root cause", "corrective actions", "severity", "evidence preservation"},
		"documentation-maintainer": {
			"README quickstart",
			"ADRs",
			"runbooks",
			"API docs",
			"setup guides",
			"changelog",
			"release notes",
		},
		"universal-skill-creator": {"domain-specific", "changelog", "anti-patterns", "compatibility", "generic copy-paste"},
	}
	for name, terms := range cases {
		rendered := renderForTest(t, name)
		lower := strings.ToLower(rendered)
		for _, term := range terms {
			if !strings.Contains(lower, strings.ToLower(term)) {
				t.Errorf("skill %q missing domain term %q", name, term)
			}
		}
	}
}

func TestEverySkillIncludesDurableSpecDrivenContext(t *testing.T) {
	for _, name := range expectedSDLCSkills {
		rendered := renderForTest(t, name)
		for _, term := range []string{"proposal/design/tasks", "spec deltas", "sync or archive", "chat-only intent"} {
			if !strings.Contains(rendered, term) {
				t.Errorf("skill %q missing spec-driven context term %q", name, term)
			}
		}
	}
}

func TestUniversalSkillCreatorForbidsGenericCopyPaste(t *testing.T) {
	rendered := renderForTest(t, "universal-skill-creator")
	want := "Never create a skill that only differs by name and description. Every generated skill must include domain-specific review scope, checklist items, decision rules, finding categories, severity guidance, output requirements, acceptance criteria, and anti-patterns. Generic operating-model text is allowed only as shared baseline, never as the complete skill body."
	if !strings.Contains(rendered, want) {
		t.Fatalf("universal-skill-creator missing exact generic-copy prohibition rule")
	}
}

func TestNoPlaceholderText(t *testing.T) {
	for _, name := range expectedSDLCSkills {
		rendered := renderForTest(t, name)
		for _, forbidden := range []string{"TODO", "TBD", "Add checks here", "More details", "Describe what this skill helps"} {
			if strings.Contains(rendered, forbidden) {
				t.Errorf("skill %q contains placeholder text %q", name, forbidden)
			}
		}
	}
}

func TestNoGeneratorSmellPhrases(t *testing.T) {
	for _, name := range expectedSDLCSkills {
		rendered := strings.ToLower(renderForTest(t, name))
		if name == "universal-skill-creator" {
			rendered = strings.ToLower(section(renderForTest(t, name), "## When to use") +
				section(renderForTest(t, name), "## Operating model") +
				section(renderForTest(t, name), "## Finding Categories") +
				section(renderForTest(t, name), "## Acceptance Criteria"))
		}
		for _, forbidden := range []string{
			"structured analysis or review",
			strings.ReplaceAll(name, "-", " ") + " evidence",
			strings.ReplaceAll(name, "-", " ") + " control",
			"review scope is explicit and skill-specific",
			"at least one concrete evidence source",
			"using generic checklist language",
		} {
			if strings.Contains(rendered, forbidden) {
				t.Errorf("skill %q contains generator-smell phrase %q", name, forbidden)
			}
		}
	}
}

func TestGenericFallbackRendersForUnknownSkill(t *testing.T) {
	rendered, err := renderSkillTemplate("some-future-skill", skillTemplateData{Name: "some-future-skill", Title: "Some Future Skill"})
	if err != nil {
		t.Fatal(err)
	}
	if rendered != "" {
		t.Errorf("expected empty string for unknown skill, got non-empty output")
	}
}

func TestPlanSkillUsesRegisteredTemplate(t *testing.T) {
	files, err := PlanSkill(SkillOptions{Name: "requirements-analyst", OutputDir: t.TempDir(), Version: "1.0.0", Platforms: []string{"codex", "claude-code"}})
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

func renderForTest(t *testing.T, name string) string {
	t.Helper()
	rendered, err := renderSkillTemplate(name, skillTemplateData{
		Name:         name,
		Title:        skillTitle(name),
		Description:  "Test description for " + name,
		Version:      testVersion,
		Since:        testSince,
		LastModified: testLastModified,
		Owner:        "platform-engineering",
		Stability:    "stable",
		License:      "MIT",
		Platforms:    []string{"codex"},
	})
	if err != nil {
		t.Fatalf("renderSkillTemplate(%q) returned error: %v", name, err)
	}
	if rendered == "" {
		t.Fatalf("renderSkillTemplate(%q) returned empty string", name)
	}
	return rendered
}

func section(markdown, heading string) string {
	start := strings.Index(markdown, heading)
	if start < 0 {
		return ""
	}
	rest := markdown[start+len(heading):]
	end := strings.Index(rest, "\n## ")
	if end >= 0 {
		return rest[:end]
	}
	return rest
}

func countBullets(s string) int {
	count := 0
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "- [ ]") {
			count++
		}
	}
	return count
}

func normalizedBody(rendered string) string {
	idx := strings.Index(rendered, "\n# ")
	if idx >= 0 {
		rendered = rendered[idx:]
	}
	rendered = strings.ToLower(rendered)
	return strings.Join(strings.Fields(rendered), " ")
}

func similarity(a, b string) float64 {
	aw := wordSet(a)
	bw := wordSet(b)
	if len(aw) == 0 || len(bw) == 0 {
		return 0
	}
	inter := 0
	for w := range aw {
		if _, ok := bw[w]; ok {
			inter++
		}
	}
	union := len(aw) + len(bw) - inter
	return float64(inter) / float64(union)
}

func wordSet(s string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, w := range strings.Fields(s) {
		if len(w) > 3 {
			out[w] = struct{}{}
		}
	}
	return out
}

func extractFrontmatterDate(rendered string) string {
	for _, line := range strings.Split(rendered, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "date:") {
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, "date:"))
			return strings.Trim(val, `"`)
		}
	}
	return ""
}
