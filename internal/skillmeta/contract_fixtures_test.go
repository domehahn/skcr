package skillmeta_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/domehahn/skcr/v2/internal/skillmeta"
)

// TestContractFixturesSuite validates all v1 schema fixtures under contracts/v1/
func TestContractFixturesSuite(t *testing.T) {
	rootDir, err := filepath.Abs("../../contracts/v1")
	if err != nil {
		t.Fatalf("resolve contracts dir: %v", err)
	}

	// 1. Valid Fixture
	validDesc := filepath.Join(rootDir, "valid", "descriptor.yaml")
	if _, err := skillmeta.LoadDescriptor(validDesc); err != nil {
		t.Fatalf("expected valid descriptor fixture to load cleanly, got: %v", err)
	}

	validContract := filepath.Join(rootDir, "valid", "contract.yaml")
	if _, err := skillmeta.LoadContract(validContract); err != nil {
		t.Fatalf("expected valid contract fixture to load cleanly, got: %v", err)
	}

	// 2. Invalid Fixtures
	invalidDesc := filepath.Join(rootDir, "invalid", "descriptor.yaml")
	if desc, err := skillmeta.LoadDescriptor(invalidDesc); err == nil {
		if errs := skillmeta.ValidateDescriptor(desc, filepath.Dir(invalidDesc)); len(errs) == 0 {
			t.Fatal("expected invalid descriptor fixture to fail validation")
		}
	} else {
		t.Fatalf("unexpected error loading descriptor: %v", err)
	}

	invalidContract := filepath.Join(rootDir, "invalid", "contract.yaml")
	if data, err := os.ReadFile(invalidContract); err == nil {
		if c, err := skillmeta.ParseContract(data); err == nil {
			if errs := skillmeta.ValidateContract(c); len(errs) == 0 {
				t.Fatal("expected invalid contract fixture with conflicting tools to fail validation")
			}
		}
	}

	// 3. Forward Compatibility Fixture
	forwardDesc := filepath.Join(rootDir, "forward-compat", "descriptor.yaml")
	if _, err := skillmeta.LoadDescriptor(forwardDesc); err != nil {
		t.Fatalf("expected forward-compatible descriptor fixture to load, got: %v", err)
	}

	// 4. Backward Compatibility Fixture
	backwardDesc := filepath.Join(rootDir, "backward-compat", "descriptor.yaml")
	if _, err := skillmeta.LoadDescriptor(backwardDesc); err != nil {
		t.Fatalf("expected backward-compatible descriptor fixture to load, got: %v", err)
	}
}
