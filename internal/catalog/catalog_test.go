package catalog

import (
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
