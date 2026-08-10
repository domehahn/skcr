package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestCategoryNamesForScope(t *testing.T) {
	semantic, err := categoryNamesForScope("semantic")
	if err != nil {
		t.Fatal(err)
	}
	if !containsCategoryName(semantic, "storage") || containsCategoryName(semantic, "cncf-runtime-cloud-native-storage") {
		t.Fatalf("semantic scope should expose curated facets only: %v", semantic)
	}

	cncfCategories, err := categoryNamesForScope("cncf")
	if err != nil {
		t.Fatal(err)
	}
	if !containsCategoryName(cncfCategories, "cncf-landscape") || !containsCategoryName(cncfCategories, "cncf-runtime-cloud-native-storage") {
		t.Fatalf("cncf scope should expose the complete official taxonomy: %v", cncfCategories)
	}
	if _, err := categoryNamesForScope("invalid"); err == nil {
		t.Fatal("expected invalid category scope to fail")
	}
}

func TestListCategoriesSemanticScope(t *testing.T) {
	cmd := newListCategoriesCommand()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"--scope", "semantic"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "storage") || strings.Contains(output.String(), "cncf-runtime-cloud-native-storage") {
		t.Fatalf("unexpected semantic category output:\n%s", output.String())
	}
}

func containsCategoryName(names []string, target string) bool {
	for _, name := range names {
		if name == target {
			return true
		}
	}
	return false
}
