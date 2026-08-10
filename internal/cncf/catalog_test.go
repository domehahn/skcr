package cncf

import (
	"strings"
	"testing"
)

func TestEmbeddedLandscapeIsCompleteAndDeterministic(t *testing.T) {
	items, err := Entries()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) < 2400 {
		t.Fatalf("expected complete CNCF Landscape snapshot, got %d unique entries", len(items))
	}
	seen := map[string]bool{}
	for _, item := range items {
		if item.Name == "" || item.SkillName == "" || len(item.SkillName) > 64 {
			t.Fatalf("invalid item: %#v", item)
		}
		if seen[item.SkillName] {
			t.Fatalf("duplicate skill name %q", item.SkillName)
		}
		seen[item.SkillName] = true
	}
	for _, expected := range []string{"cncf-ansible-reviewer", "cncf-kubernetes-reviewer", "cncf-prometheus-reviewer"} {
		if !seen[expected] {
			t.Errorf("missing expected Landscape skill %q", expected)
		}
	}
	if !strings.HasPrefix(SnapshotDigest(), "sha256:") {
		t.Fatalf("unexpected snapshot digest %q", SnapshotDigest())
	}
}

func TestSkillCategoriesIncludeTopLevelAndSubcategories(t *testing.T) {
	categories := SkillCategories()
	for _, name := range []string{"cncf", "cncf-landscape", "cncf-graduated", "cncf-incubating", "cncf-sandbox", "cncf-archived", "cncf-provisioning", "cncf-runtime", "cncf-members", "cncf-provisioning-automation-and-configuration"} {
		if len(categories[name]) == 0 {
			t.Errorf("missing or empty category %q", name)
		}
	}
	if len(categories["cncf-landscape"]) != len(MustEntries()) {
		t.Fatalf("cncf-landscape category has %d skills, want %d", len(categories["cncf-landscape"]), len(MustEntries()))
	}
	if len(categories["cncf"]) >= len(categories["cncf-landscape"]) || len(categories["cncf"]) < 200 {
		t.Fatalf("cncf should contain active official projects, got %d versus %d Landscape entries", len(categories["cncf"]), len(categories["cncf-landscape"]))
	}
}

func TestEveryLandscapeEntryHasSemanticCategory(t *testing.T) {
	categories := SkillCategories()
	semanticNames := SemanticCategoryNames()
	for _, item := range MustEntries() {
		found := false
		for _, category := range semanticNames {
			if containsString(categories[category], item.SkillName) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Landscape skill %q has no semantic category", item.SkillName)
		}
	}
	for category, expected := range map[string]string{
		"security":                "cncf-trivy-reviewer",
		"storage":                 "cncf-ceph-reviewer",
		"runtime-containers":      "cncf-containerd-reviewer",
		"networking-service-mesh": "cncf-cilium-reviewer",
		"observability":           "cncf-prometheus-reviewer",
		"databases":               "cncf-postgresql-reviewer",
	} {
		if !containsString(categories[category], expected) {
			t.Errorf("semantic category %q missing %q", category, expected)
		}
	}
}

func TestBroadTechnologyFacetsExcludeAdjacentVendorGroups(t *testing.T) {
	categories := SkillCategories()
	if containsString(categories["storage"], "cncf-postgresql-reviewer") {
		t.Fatal("database products should be selected through databases, not storage")
	}
	if containsString(categories["kubernetes"], "cncf-syseleven-kcsp-reviewer") {
		t.Fatal("Kubernetes service providers should be selected through service-providers")
	}
	if containsString(categories["kubernetes"], "cncf-agileops-kcntp-reviewer") {
		t.Fatal("Kubernetes training should be selected through training-certification")
	}
}
