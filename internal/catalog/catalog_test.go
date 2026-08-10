package catalog

import (
	"strings"
	"testing"

	"github.com/domehahn/skcr/internal/scaffold"
)

func TestCoreSkillsMatchSDLCSkillNames(t *testing.T) {
	sdlc := scaffold.SDLCSkillNames
	if len(CoreSkills) != len(sdlc) {
		t.Fatalf("catalog.CoreSkills has %d entries, scaffold.SDLCSkillNames has %d", len(CoreSkills), len(sdlc))
	}
	for i, want := range sdlc {
		if CoreSkills[i] != want {
			t.Errorf("position %d: catalog.CoreSkills[%d]=%q, scaffold.SDLCSkillNames[%d]=%q", i, i, CoreSkills[i], i, want)
		}
	}
	for _, name := range sdlc {
		if _, ok := SkillDescriptions[name]; !ok {
			t.Errorf("catalog.SkillDescriptions missing entry for %q", name)
		}
	}
}

func TestSkillTitleAndDescription(t *testing.T) {
	if got := SkillTitle("security-reviewer"); got != "Security Reviewer" {
		t.Fatalf("SkillTitle mismatch: %q", got)
	}
	if got := SkillTitle("double--dash"); got != "Double  Dash" {
		t.Fatalf("SkillTitle preserves empty segment spacing, got: %q", got)
	}

	known := SkillDescription("security-reviewer")
	if known == "" || known == "Reusable agent skill for Security Reviewer tasks." {
		t.Fatalf("expected known description, got %q", known)
	}

	unknown := SkillDescription("custom-skill")
	want := "Reusable agent skill for Custom Skill tasks."
	if unknown != want {
		t.Fatalf("unexpected fallback description: %q", unknown)
	}
}

func TestSkillCategories(t *testing.T) {
	names := CategoryNames()
	if len(names) == 0 {
		t.Fatal("expected built-in categories")
	}

	canonical, err := NormalizeCategory("DORA/VAIT")
	if err != nil {
		t.Fatal(err)
	}
	if canonical != "dora-vait" {
		t.Fatalf("unexpected category normalization: %q", canonical)
	}

	skills, err := SkillsForCategory("dora")
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 8 {
		t.Fatalf("expected 8 DORA skills, got %d: %#v", len(skills), skills)
	}
	if skills[0] != "dora-readiness-reviewer" {
		t.Fatalf("unexpected first DORA skill: %q", skills[0])
	}

	agentic, err := SkillsForCategory("agent-security")
	if err != nil {
		t.Fatal(err)
	}
	if len(agentic) != 6 || agentic[0] != "agent-containment-reviewer" ||
		agentic[5] != "security-invariant-test-engineer" {
		t.Fatalf("unexpected agentic-security skills: %#v", agentic)
	}

	payments, err := SkillsForCategory("stripe")
	if err != nil {
		t.Fatal(err)
	}
	if len(payments) != 16 || payments[0] != "payment-integration-engineer" ||
		payments[15] != "adyen-integration-engineer" {
		t.Fatalf("unexpected payment skills: %#v", payments)
	}

	if _, err := SkillsForCategory("not-a-category"); err == nil {
		t.Fatal("expected unknown category error")
	}

	for category, expected := range map[string]string{
		"languages":      "python-reviewer",
		"frameworks":     "quarkus-reviewer",
		"infrastructure": "opentofu-reviewer",
		"cncf":           "cncf-prometheus-reviewer",
		"cncf-runtime":   "cncf-containerd-reviewer",
	} {
		skills, err := SkillsForCategory(category)
		if err != nil {
			t.Fatal(err)
		}
		if !containsCatalogSkill(skills, expected) {
			t.Errorf("category %q missing %q", category, expected)
		}
	}

	cncfSkills, err := SkillsForCategory("landscape")
	if err != nil {
		t.Fatal(err)
	}
	if len(cncfSkills) < 2400 {
		t.Fatalf("expected complete CNCF Landscape category, got %d skills", len(cncfSkills))
	}
}

func TestCoreSkillsAreUniqueAndGeneratedNamesAreBounded(t *testing.T) {
	seen := map[string]bool{}
	for _, name := range CoreSkills {
		if seen[name] {
			t.Fatalf("duplicate core skill %q", name)
		}
		seen[name] = true
		if strings.HasPrefix(name, "cncf-") && len(name) > 64 {
			t.Fatalf("CNCF skill name exceeds 64 characters: %q", name)
		}
	}
}

func TestEveryCoreSkillHasSemanticCategory(t *testing.T) {
	semantic := SemanticCategoryNames()
	for _, skill := range CoreSkills {
		found := false
		for _, category := range semantic {
			if containsCatalogSkill(SkillCategories[category], skill) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("core skill %q has no semantic category", skill)
		}
	}
}

func TestSemanticCategoriesSplitLandscapeByDomain(t *testing.T) {
	cases := map[string]string{
		"security":                   "cncf-trivy-reviewer",
		"storage":                    "cncf-ceph-reviewer",
		"runtime-containers":         "cncf-containerd-reviewer",
		"networking-service-mesh":    "cncf-cilium-reviewer",
		"observability":              "cncf-prometheus-reviewer",
		"organizations-members":      "cncf-red-hat-member-reviewer",
		"service-providers":          "cncf-syseleven-kcsp-reviewer",
		"infrastructure-as-code":     "terraform-reviewer",
		"backup-disaster-recovery":   "backup-restore-reviewer",
		"release-feature-management": "release-readiness-reviewer",
	}
	for category, skill := range cases {
		if !containsCatalogSkill(SkillCategories[category], skill) {
			t.Errorf("category %q missing representative skill %q", category, skill)
		}
	}
	if len(SkillCategories["cncf"]) >= 300 {
		t.Fatalf("cncf should contain only active official projects, got %d skills", len(SkillCategories["cncf"]))
	}
	if len(SkillCategories["cncf-landscape"]) < 2400 {
		t.Fatalf("cncf-landscape should retain every Landscape entry, got %d", len(SkillCategories["cncf-landscape"]))
	}
	if len(SkillCategories["storage"]) >= 120 {
		t.Fatalf("storage should not absorb the separate database category, got %d skills", len(SkillCategories["storage"]))
	}
	if len(SkillCategories["kubernetes"]) >= 300 {
		t.Fatalf("kubernetes should not absorb provider and training directories, got %d skills", len(SkillCategories["kubernetes"]))
	}
}

func containsCatalogSkill(skills []string, target string) bool {
	for _, skill := range skills {
		if skill == target {
			return true
		}
	}
	return false
}
