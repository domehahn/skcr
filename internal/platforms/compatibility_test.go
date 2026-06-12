package platforms

import "testing"

func TestCompatibilityMatrixIsExplicitAboutValidationState(t *testing.T) {
	if len(CompatibilityMatrix) == 0 {
		t.Fatal("compatibility matrix must not be empty")
	}
	seen := map[string]struct{}{}
	for _, entry := range CompatibilityMatrix {
		if entry.Name == "" {
			t.Fatal("platform name must not be empty")
		}
		if entry.MinVersion == "" {
			t.Fatalf("platform %s must declare a minimum version or unknown", entry.Name)
		}
		if entry.MinVersion == "unknown" && entry.Status != "unverified" {
			t.Fatalf("platform %s with unknown version must be unverified, got %q", entry.Name, entry.Status)
		}
		if entry.MinVersion != "unknown" && entry.Status != "verified" {
			t.Fatalf("platform %s with concrete version must be verified, got %q", entry.Name, entry.Status)
		}
		if _, ok := seen[entry.Name]; ok {
			t.Fatalf("duplicate platform compatibility entry: %s", entry.Name)
		}
		seen[entry.Name] = struct{}{}
		if got, ok := MinVersion(entry.Name); !ok || got != entry.MinVersion {
			t.Fatalf("MinVersion(%q) = %q, %v; want %q, true", entry.Name, got, ok, entry.MinVersion)
		}
	}
}

func TestMinVersionsForPreservesMatrixOrder(t *testing.T) {
	got := MinVersionsFor([]string{"gitlab-duo", "codex"})
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %#v", got)
	}
	if got[0].Name != "codex" || got[1].Name != "gitlab-duo" {
		t.Fatalf("expected matrix order, got %#v", got)
	}
}

func TestToolCapabilitiesCoverOpenSpecStyleSkillAndCommandSurfaces(t *testing.T) {
	for _, name := range []string{"codex", "github-copilot", "kiro", "junie", "gemini-cli", "antigravity", "cline", "amazon-q", "qoder", "qwen"} {
		capability, ok := CapabilityFor(name)
		if !ok {
			t.Fatalf("missing tool capability for %s", name)
		}
		if capability.SkillPathPattern == "" {
			t.Fatalf("tool capability %s must declare a skill path pattern", name)
		}
		if capability.Delivery == DeliveryBoth && capability.CommandPathPattern == "" {
			t.Fatalf("tool capability %s with delivery=both must declare a command path pattern", name)
		}
		if capability.Status != "unverified" {
			t.Fatalf("tool capability %s should remain unverified until validated, got %q", name, capability.Status)
		}
	}
}
