package platforms

import "testing"

func TestCompatibilityMatrixIsProductionReady(t *testing.T) {
	if len(CompatibilityMatrix) == 0 {
		t.Fatal("compatibility matrix must not be empty")
	}
	seen := map[string]struct{}{}
	for _, entry := range CompatibilityMatrix {
		if entry.Name == "" {
			t.Fatal("platform name must not be empty")
		}
		if entry.MinVersion == "" || entry.MinVersion == "unknown" {
			t.Fatalf("platform %s must have a concrete minimum version", entry.Name)
		}
		if entry.Status != "verified" {
			t.Fatalf("platform %s must be verified for production built-ins, got %q", entry.Name, entry.Status)
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
