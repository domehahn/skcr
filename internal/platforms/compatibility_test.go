package platforms

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
		if entry.MinVersion != "unknown" && entry.Evidence == "" {
			t.Fatalf("platform %s with concrete version must declare evidence", entry.Name)
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

func TestCompatibilityEvidenceRoundTrip(t *testing.T) {
	root := t.TempDir()
	evidence := filepath.Join(root, "docs", "compat", "codex-0.51.0.md")
	if err := os.MkdirAll(filepath.Dir(evidence), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(evidence, []byte("# Codex compatibility\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	entry := CompatibilityEntry{
		Name:       "codex",
		MinVersion: "0.51.0",
		Status:     "verified",
		Source:     "local-evidence",
		Evidence:   "docs/compat/codex-0.51.0.md",
		Validated:  "2026-06-12",
	}
	if err := SaveEvidence(root, entry); err != nil {
		t.Fatal(err)
	}
	matrix, err := LoadMatrix(root)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range matrix {
		if item.Name == "codex" {
			found = true
			if item.MinVersion != "0.51.0" || item.Status != "verified" || item.Evidence == "" {
				t.Fatalf("unexpected codex entry: %#v", item)
			}
		}
	}
	if !found {
		t.Fatal("codex entry not found")
	}
	payload, err := os.ReadFile(filepath.Join(root, "agentic.compatibility.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "unknown") {
		t.Fatalf("compatibility evidence file should store verified overrides only: %s", payload)
	}
}

func TestValidateEvidenceRejectsConcreteVersionWithoutEvidence(t *testing.T) {
	err := ValidateEvidenceEntry(t.TempDir(), CompatibilityEntry{Name: "codex", MinVersion: "0.51.0", Status: "verified"})
	if err == nil {
		t.Fatal("expected missing evidence error")
	}
}

func TestValidateEvidenceRejectsPathsOutsideTarget(t *testing.T) {
	root := t.TempDir()
	for _, evidence := range []string{"/tmp/codex.md", "../codex.md"} {
		err := ValidateEvidenceEntry(root, CompatibilityEntry{
			Name:       "codex",
			MinVersion: "0.51.0",
			Status:     "verified",
			Evidence:   evidence,
			Validated:  "2026-06-12",
		})
		if err == nil {
			t.Fatalf("expected evidence path %q to be rejected", evidence)
		}
	}
}

func TestValidateEvidenceRejectsInvalidValidatedDate(t *testing.T) {
	root := t.TempDir()
	evidence := filepath.Join(root, "docs", "compat", "codex.md")
	if err := os.MkdirAll(filepath.Dir(evidence), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(evidence, []byte("# evidence\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := ValidateEvidenceEntry(root, CompatibilityEntry{
		Name:       "codex",
		MinVersion: "0.51.0",
		Status:     "verified",
		Evidence:   "docs/compat/codex.md",
		Validated:  "12.06.2026",
	})
	if err == nil {
		t.Fatal("expected invalid validated date error")
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

func TestValidateEvidenceEntryAllErrorPaths(t *testing.T) {
	root := t.TempDir()

	// Empty name.
	if err := ValidateEvidenceEntry(root, CompatibilityEntry{Name: "", MinVersion: "1.0", Status: "verified"}); err == nil {
		t.Fatal("expected error for empty name")
	}

	// Empty min_version.
	if err := ValidateEvidenceEntry(root, CompatibilityEntry{Name: "codex", MinVersion: "", Status: "verified"}); err == nil {
		t.Fatal("expected error for empty min_version")
	}

	// unknown min_version with non-unverified status.
	if err := ValidateEvidenceEntry(root, CompatibilityEntry{Name: "codex", MinVersion: "unknown", Status: "verified"}); err == nil {
		t.Fatal("expected error for unknown version with verified status")
	}

	// unknown min_version and unverified status → ok.
	if err := ValidateEvidenceEntry(root, CompatibilityEntry{Name: "codex", MinVersion: "unknown", Status: "unverified"}); err != nil {
		t.Fatalf("unknown+unverified should succeed: %v", err)
	}

	// Concrete version without verified status.
	if err := ValidateEvidenceEntry(root, CompatibilityEntry{Name: "codex", MinVersion: "1.0", Status: "unverified"}); err == nil {
		t.Fatal("expected error for concrete version with unverified status")
	}

	// Evidence file declared but does not exist.
	if err := ValidateEvidenceEntry(root, CompatibilityEntry{
		Name: "codex", MinVersion: "1.0", Status: "verified",
		Evidence: "docs/missing.md", Validated: "2026-06-12",
	}); err == nil {
		t.Fatal("expected error for missing evidence file")
	}

	// Missing validated date.
	evFile := filepath.Join(root, "ev.md")
	if err := os.WriteFile(evFile, []byte("# ev\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateEvidenceEntry(root, CompatibilityEntry{
		Name: "codex", MinVersion: "1.0", Status: "verified",
		Evidence: "ev.md",
	}); err == nil {
		t.Fatal("expected error for missing validated date")
	}

	// All valid.
	if err := ValidateEvidenceEntry(root, CompatibilityEntry{
		Name: "codex", MinVersion: "1.0", Status: "verified",
		Evidence: "ev.md", Validated: "2026-06-12",
	}); err != nil {
		t.Fatalf("fully valid entry should succeed: %v", err)
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
