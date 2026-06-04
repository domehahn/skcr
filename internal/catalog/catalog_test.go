package catalog

import "testing"

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
