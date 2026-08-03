package asps

import "testing"

func TestPinnedCatalog(t *testing.T) {
	if len(Domains()) != 15 || len(Properties()) != 120 {
		t.Fatalf("unexpected ASPS catalog size: %d/%d", len(Domains()), len(Properties()))
	}
	property, ok := FindProperty("ASP-08.03")
	if !ok || property.Name != "Delegation Monotonicity" {
		t.Fatalf("unexpected property: %#v", property)
	}
	if !KnownProfile("asps-core@1.0") || len(ProfileProperties("asps-mcp@1.0")) == 0 {
		t.Fatal("expected pinned profiles")
	}
}
