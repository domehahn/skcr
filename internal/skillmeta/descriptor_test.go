package skillmeta

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/domehahn/sklib/spec"
	"gopkg.in/yaml.v3"
)

func writeArtifactFixture(t *testing.T) (string, Descriptor, Contract) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "example")
	if err := os.MkdirAll(filepath.Join(dir, "evals"), 0o755); err != nil {
		t.Fatal(err)
	}
	descriptor := NewDescriptor("example", "1.0.0", "Produce an example result.", "MIT", []string{"owner"}, []spec.Platform{spec.PlatformCodex})
	contract := NewContract()
	writeYAML(t, filepath.Join(dir, "skill.yaml"), descriptor)
	writeYAML(t, filepath.Join(dir, "contract.yaml"), contract)
	writeYAML(t, filepath.Join(dir, "evals", "baseline.yaml"), NewBaselineEval())
	return dir, descriptor, contract
}

func writeYAML(t *testing.T, path string, value any) {
	t.Helper()
	data, err := yaml.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSecureDefaultsAndValidArtifact(t *testing.T) {
	dir, descriptor, contract := writeArtifactFixture(t)
	if errs := ValidateDirectory(dir); len(errs) != 0 {
		t.Fatalf("valid artifact errors: %v", errs)
	}
	if descriptor.Contract.File != "contract.yaml" || descriptor.Evals.Directory != "evals" {
		t.Fatalf("unexpected references: %#v", descriptor)
	}
	allowed := contract.Capabilities.Allowed
	if len(allowed.Repository.Write) != 0 || len(allowed.Filesystem.Write) != 0 ||
		len(allowed.Network.Allow) != 0 || *allowed.Process.Execute ||
		*allowed.Secrets.Read || len(contract.Tools.Allow) != 0 {
		t.Fatal("new contract is not least privilege")
	}
}

func TestDescriptorReferenceValidation(t *testing.T) {
	dir, descriptor, _ := writeArtifactFixture(t)
	tests := map[string]struct {
		mutate func(*Descriptor)
		want   string
	}{
		"missing goal":       {func(d *Descriptor) { d.Goal = nil }, "goal is required"},
		"blank objective":    {func(d *Descriptor) { d.Goal.Objective = " " }, "goal.objective"},
		"missing contract":   {func(d *Descriptor) { d.Contract = nil }, "contract reference is required"},
		"contract traversal": {func(d *Descriptor) { d.Contract.File = "../../outside" }, "contract.file"},
		"eval traversal":     {func(d *Descriptor) { d.Evals.Directory = "../../outside" }, "evals.directory"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			d := descriptor
			goal := *descriptor.Goal
			contractRef := *descriptor.Contract
			evalsRef := *descriptor.Evals
			d.Goal, d.Contract, d.Evals = &goal, &contractRef, &evalsRef
			tc.mutate(&d)
			errs := strings.Join(ValidateDescriptor(d, dir), "\n")
			if !strings.Contains(errs, tc.want) {
				t.Fatalf("expected %q, got %q", tc.want, errs)
			}
		})
	}
}

func TestContractValidationFailures(t *testing.T) {
	_, _, base := writeArtifactFixture(t)
	tests := map[string]struct {
		mutate func(*Contract)
		want   string
	}{
		"tool conflict": {func(c *Contract) {
			c.Tools.Allow = []string{"repo.write"}
			c.Tools.Deny = []string{"repo.write"}
		}, "both allow and deny"},
		"negative limit": {func(c *Contract) {
			n := -1
			c.Limits.MaxRuntimeSeconds = &n
		}, "zero or greater"},
		"wildcard": {func(c *Contract) { c.Tools.Allow = []string{"*"} }, "wildcard"},
		"required outside allowed": {func(c *Contract) {
			c.Capabilities.Required.Repository.Read = []string{"source_files"}
		}, "outside allowed"},
		"unsupported schema": {func(c *Contract) { c.SchemaVersion = "99" }, "unsupported"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c := base
			tc.mutate(&c)
			errs := strings.Join(ValidateContract(c), "\n")
			if !strings.Contains(errs, tc.want) {
				t.Fatalf("expected %q, got %q", tc.want, errs)
			}
		})
	}
}

func TestMissingAndNullLimitSemantics(t *testing.T) {
	valid, err := yaml.Marshal(NewContract())
	if err != nil {
		t.Fatal(err)
	}
	contract, err := ParseContract(valid)
	if err != nil || len(ValidateContract(contract)) != 0 {
		t.Fatalf("explicit null limits should be valid: %v %v", err, ValidateContract(contract))
	}
	missing := strings.Replace(string(valid), "    max_tool_calls: null\n", "", 1)
	contract, err = ParseContract([]byte(missing))
	if err != nil {
		t.Fatal(err)
	}
	if errs := strings.Join(ValidateContract(contract), "\n"); !strings.Contains(errs, "max_tool_calls is required") {
		t.Fatalf("missing limit should fail, got %q", errs)
	}
}

func TestLegacyAndEmbeddedV2Compatibility(t *testing.T) {
	legacy := []byte("name: old-skill\nversion: 1.0.0\ndescription: old\nentrypoint: SKILL.md\ncompatible_with: []\nunknown_legacy_field: retained\n")
	d, err := ParseDescriptor(legacy)
	if err != nil || d.SchemaVersion != "" {
		t.Fatalf("legacy descriptor should parse without new semantics: %#v %v", d, err)
	}
	current := []byte("schema_version: \"2\"\nname: example\nversion: 1.0.0\ndescription: x\nentrypoint: SKILL.md\ncompatible_with: []\nunknown_security_field: true\n")
	if _, err := ParseDescriptor(current); err == nil {
		t.Fatal("current schema should reject unknown fields")
	}
}

func TestArtifactReferenceSymlinkEscape(t *testing.T) {
	dir, descriptor, _ := writeArtifactFixture(t)
	outside := filepath.Join(t.TempDir(), "outside.yaml")
	if err := os.WriteFile(outside, []byte("schema_version: \"1\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "escaped-contract.yaml")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	descriptor.Contract.File = "escaped-contract.yaml"
	if errs := strings.Join(ValidateDescriptor(descriptor, dir), "\n"); !strings.Contains(errs, "contract.file") {
		t.Fatalf("symlink escape should fail, got %q", errs)
	}
	outsideDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outsideDir, "outside.yaml"), []byte("schema_version: \"1\"\nscenarios: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	evalLink := filepath.Join(dir, "escaped-evals")
	if err := os.Symlink(outsideDir, evalLink); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	descriptor.Evals.Directory = "escaped-evals"
	if errs := strings.Join(ValidateDescriptor(descriptor, dir), "\n"); !strings.Contains(errs, "evals.directory") {
		t.Fatalf("eval symlink escape should fail, got %q", errs)
	}
}

func TestEvalValidationFailures(t *testing.T) {
	dir, descriptor, contract := writeArtifactFixture(t)
	path := filepath.Join(dir, "evals", "baseline.yaml")
	if err := os.WriteFile(path, []byte("schema_version: ["), 0o644); err != nil {
		t.Fatal(err)
	}
	if errs := strings.Join(ValidateEvalDirectory(dir, "evals", descriptor, &contract), "\n"); !strings.Contains(errs, "malformed eval") {
		t.Fatalf("malformed eval not rejected: %q", errs)
	}
	unknown := NewBaselineEval()
	unknown.Scenarios[0].Assertions.Capabilities.MustNotUse = append(unknown.Scenarios[0].Assertions.Capabilities.MustNotUse, "provider.magic")
	writeYAML(t, path, unknown)
	if errs := strings.Join(ValidateEvalDirectory(dir, "evals", descriptor, &contract), "\n"); !strings.Contains(errs, "unknown capability") {
		t.Fatalf("unknown assertion not rejected: %q", errs)
	}
	unsupported := NewBaselineEval()
	unsupported.SchemaVersion = "99"
	writeYAML(t, path, unsupported)
	if errs := strings.Join(ValidateEvalDirectory(dir, "evals", descriptor, &contract), "\n"); !strings.Contains(errs, "unsupported eval") {
		t.Fatalf("unsupported eval schema not rejected: %q", errs)
	}
}

func TestNormalizationDigestAndDiff(t *testing.T) {
	left := NewContract()
	left.Tools.Allow = []string{"bar", "foo", "bar"}
	right := NewContract()
	right.Tools.Allow = []string{"foo", "bar"}
	leftDigest, _ := ContractDigest(left)
	rightDigest, _ := ContractDigest(right)
	if leftDigest != rightDigest {
		t.Fatalf("equivalent contracts have different digests: %s != %s", leftDigest, rightDigest)
	}
	expanded := NewContract()
	expanded.Capabilities.Allowed.Network.Allow = []string{"api.example.com"}
	diff := DiffContracts(NewContract(), expanded)
	if diff.Classification != ImpactExpansion {
		t.Fatalf("expected expansion, got %#v", diff)
	}
	tightened := expanded
	tightened.Capabilities.Allowed.Network.Allow = []string{}
	if diff := DiffContracts(expanded, tightened); diff.Classification != ImpactNarrowing {
		t.Fatalf("expected narrowing, got %#v", diff)
	}
	requiredOnly := NewContract()
	requiredOnly.Capabilities.Required.Repository.Read = []string{"source_files"}
	requiredOnly.Capabilities.Allowed.Repository.Read = []string{"source_files"}
	allowedBaseline := NewContract()
	allowedBaseline.Capabilities.Allowed.Repository.Read = []string{"source_files"}
	if diff := DiffContracts(allowedBaseline, requiredOnly); diff.Classification != ImpactNone {
		t.Fatalf("required-only change should not expand allowed boundary: %#v", diff)
	}
}

func TestPublicSchemasAreValidJSONAndDeclareVersions(t *testing.T) {
	root := filepath.Join("..", "..", "schemas")
	for file, version := range map[string]string{
		"skill-v2.schema.json": "2", "contract-v1.schema.json": "1", "eval-v1.schema.json": "1",
	} {
		data, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatal(err)
		}
		var schema map[string]any
		if err := json.Unmarshal(data, &schema); err != nil {
			t.Fatalf("%s is not valid JSON: %v", file, err)
		}
		if !strings.Contains(string(data), `"const": "`+version+`"`) || schema["$schema"] == nil {
			t.Fatalf("%s does not declare schema version %s", file, version)
		}
	}
}

func TestGoldenContractFixtures(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "contracts")
	for _, kind := range []string{"valid", "invalid"} {
		entries, err := os.ReadDir(filepath.Join(root, kind))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			t.Run(kind+"/"+entry.Name(), func(t *testing.T) {
				contract, err := LoadContract(filepath.Join(root, kind, entry.Name(), "contract.yaml"))
				if err != nil {
					t.Fatal(err)
				}
				errs := ValidateContract(contract)
				if kind == "valid" && len(errs) > 0 {
					t.Fatalf("valid fixture failed: %v", errs)
				}
				if kind == "invalid" && len(errs) == 0 {
					t.Fatal("invalid fixture unexpectedly passed")
				}
			})
		}
	}
}
