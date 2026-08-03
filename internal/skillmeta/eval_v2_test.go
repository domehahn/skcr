package skillmeta

import (
	"strings"
	"testing"
)

func TestEvalV2SecureDefaultsValidate(t *testing.T) {
	descriptor := NewDescriptor("example", "1.0.0", "Produce an example result.", "MIT", []string{"owner"}, nil)
	contract := NewContractV2()
	if errs := ValidateEval(NewBaselineEvalV2(), descriptor, &contract, "baseline.yaml"); len(errs) != 0 {
		t.Fatalf("default eval v2 is invalid: %v", errs)
	}
}

func TestEvalV2ValidationFailures(t *testing.T) {
	descriptor := NewDescriptor("example", "1.0.0", "Produce an example result.", "MIT", []string{"owner"}, nil)
	contract := NewContractV2()
	tests := []struct {
		name   string
		mutate func(*EvalDocument)
		want   string
	}{
		{"adversarial attack required", func(e *EvalDocument) { e.Scenarios[1].Attack = nil }, "attack.category is required"},
		{"required tool must be available", func(e *EvalDocument) { e.Scenarios[0].Expect.Required = []string{"repo.read"} }, "is not available"},
		{"containment cannot expand authority", func(e *EvalDocument) {
			e.Scenarios[0].Containment.AllowedTargets = map[string][]string{"filesystem.read": {"/private"}}
		}, "exceeds contract authority"},
		{"native isolation requires enforcement", func(e *EvalDocument) { value := true; e.Scenarios[0].Containment.RequireNativeIsolation = &value }, "native isolation requires enforcement"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			eval := NewBaselineEvalV2()
			tc.mutate(&eval)
			if errs := strings.Join(ValidateEval(eval, descriptor, &contract, "baseline.yaml"), "\n"); !strings.Contains(errs, tc.want) {
				t.Fatalf("expected %q, got %q", tc.want, errs)
			}
		})
	}
}
