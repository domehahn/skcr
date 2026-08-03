package skillmeta

import (
	"gopkg.in/yaml.v3"
	"strings"
	"testing"
)

func TestDefaultAssuranceRequirementsValidate(t *testing.T) {
	if errs := ValidateAssurance(NewAssuranceDocument(), []EvalDocument{NewBaselineEvalV2()}); len(errs) != 0 {
		t.Fatalf("default assurance invalid: %v", errs)
	}
	if len(RequiredASPSProperties(NewAssuranceDocument())) == 0 {
		t.Fatal("profile did not expand")
	}
}

func TestAssuranceRejectsClaimsAndUnsupportedRequirements(t *testing.T) {
	data, err := yaml.Marshal(NewAssuranceDocument())
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, []byte("verified: true\n")...)
	if _, err := ParseAssurance(data); err == nil {
		t.Fatal("assurance claims must fail strict parsing")
	}
	document := NewAssuranceDocument()
	document.ASPS.RequiredProperties = []string{"ASP-99.99"}
	document.Assurance.MinimumRequestedLevel = "A9"
	errs := strings.Join(ValidateAssurance(document, []EvalDocument{NewBaselineEvalV2()}), "\n")
	for _, want := range []string{"unknown property", "A1 through A5"} {
		if !strings.Contains(errs, want) {
			t.Fatalf("expected %q, got %q", want, errs)
		}
	}
}
