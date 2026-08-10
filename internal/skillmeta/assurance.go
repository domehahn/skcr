package skillmeta

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/domehahn/skcr/v2/internal/asps"
)

const AssuranceSchemaVersion = "1"

type AssuranceDocument struct {
	SchemaVersion  string                `yaml:"schema_version" json:"schema_version"`
	ASPS           ASPSRequirements      `yaml:"asps" json:"asps"`
	Assurance      AssuranceRequirements `yaml:"assurance" json:"assurance"`
	SecurityReview SecurityReview        `yaml:"security_review" json:"security_review"`
}
type ASPSRequirements struct {
	Specification      string   `yaml:"specification" json:"specification"`
	Version            string   `yaml:"version" json:"version"`
	RequiredProfiles   []string `yaml:"required_profiles" json:"required_profiles"`
	RequiredProperties []string `yaml:"required_properties" json:"required_properties"`
}
type AssuranceRequirements struct {
	MinimumRequestedLevel string                  `yaml:"minimum_requested_level" json:"minimum_requested_level"`
	Requirements          AssuranceRequirementSet `yaml:"requirements" json:"requirements"`
}
type AssuranceRequirementSet struct {
	BehavioralEval  *bool `yaml:"behavioral_eval" json:"behavioral_eval"`
	AdversarialEval *bool `yaml:"adversarial_eval" json:"adversarial_eval"`
	Containment     *bool `yaml:"containment" json:"containment"`
	NativeIsolation *bool `yaml:"native_isolation" json:"native_isolation"`
	Provenance      *bool `yaml:"provenance" json:"provenance"`
}
type SecurityReview struct {
	ExpansionApprovals []ExpansionApproval `yaml:"expansion_approvals" json:"expansion_approvals"`
}
type ExpansionApproval struct {
	ContractDigest string `yaml:"contract_digest" json:"contract_digest"`
	ApprovedBy     string `yaml:"approved_by" json:"approved_by"`
	ApprovedAt     string `yaml:"approved_at" json:"approved_at"`
	Justification  string `yaml:"justification" json:"justification"`
}

func NewAssuranceDocument() AssuranceDocument {
	return AssuranceDocument{SchemaVersion: AssuranceSchemaVersion,
		ASPS:           ASPSRequirements{Specification: "ASPS", Version: "1.0", RequiredProfiles: []string{"asps-core@1.0"}, RequiredProperties: []string{}},
		Assurance:      AssuranceRequirements{MinimumRequestedLevel: "A1", Requirements: AssuranceRequirementSet{BehavioralEval: boolPtr(true), AdversarialEval: boolPtr(true), Containment: boolPtr(true), NativeIsolation: boolPtr(true), Provenance: boolPtr(true)}},
		SecurityReview: SecurityReview{ExpansionApprovals: []ExpansionApproval{}},
	}
}

func ParseAssurance(data []byte) (AssuranceDocument, error) {
	var value AssuranceDocument
	err := parseStrict(data, &value)
	return value, err
}
func LoadAssurance(path string) (AssuranceDocument, error) {
	var value AssuranceDocument
	err := loadStrict(path, &value)
	return value, err
}

func ValidateAssurance(document AssuranceDocument, evals []EvalDocument) []string {
	if document.SchemaVersion != AssuranceSchemaVersion {
		return []string{fmt.Sprintf("unsupported assurance schema_version %q", document.SchemaVersion)}
	}
	var errs []string
	if document.ASPS.Specification != "ASPS" {
		errs = append(errs, `asps.specification must be "ASPS"`)
	}
	if document.ASPS.Version != "1.0" {
		errs = append(errs, `asps.version must be "1.0"`)
	}
	errs = append(errs, validateStringList("asps.required_profiles", document.ASPS.RequiredProfiles, false)...)
	errs = append(errs, validateStringList("asps.required_properties", document.ASPS.RequiredProperties, false)...)
	for _, profile := range document.ASPS.RequiredProfiles {
		if !asps.KnownProfile(profile) {
			errs = append(errs, fmt.Sprintf("asps.required_profiles contains unknown profile %q", profile))
		}
	}
	for _, property := range document.ASPS.RequiredProperties {
		if !asps.KnownProperty(property) {
			errs = append(errs, fmt.Sprintf("asps.required_properties contains unknown property %q", property))
		}
	}
	if !map[string]bool{"A1": true, "A2": true, "A3": true, "A4": true, "A5": true}[document.Assurance.MinimumRequestedLevel] {
		errs = append(errs, "assurance.minimum_requested_level must be A1 through A5")
	}
	requirements := document.Assurance.Requirements
	for name, value := range map[string]*bool{"behavioral_eval": requirements.BehavioralEval, "adversarial_eval": requirements.AdversarialEval, "containment": requirements.Containment, "native_isolation": requirements.NativeIsolation, "provenance": requirements.Provenance} {
		if value == nil {
			errs = append(errs, "assurance.requirements."+name+" is required and must be boolean")
		}
	}
	behavioral, adversarial, containment, native := false, false, false, false
	for _, eval := range evals {
		for _, scenario := range eval.Scenarios {
			if scenario.Type == "behavioral" || scenario.Category == "baseline" {
				behavioral = true
			}
			if scenario.Type == "adversarial" || scenario.Category == "adversarial" {
				adversarial = true
			}
			if scenario.Containment != nil && enabled(scenario.Containment.Required) {
				containment = true
			}
			if scenario.Containment != nil && enabled(scenario.Containment.RequireNativeIsolation) {
				native = true
			}
		}
	}
	for _, check := range []struct {
		name     string
		required *bool
		present  bool
	}{{"behavioral_eval", requirements.BehavioralEval, behavioral}, {"adversarial_eval", requirements.AdversarialEval, adversarial}, {"containment", requirements.Containment, containment}, {"native_isolation", requirements.NativeIsolation, native}} {
		if enabled(check.required) && !check.present {
			errs = append(errs, "assurance requirement "+check.name+" has no supporting eval declaration")
		}
	}
	seen := map[string]bool{}
	for i, approval := range document.SecurityReview.ExpansionApprovals {
		prefix := fmt.Sprintf("security_review.expansion_approvals[%d]", i)
		if !sha256DigestRE.MatchString(approval.ContractDigest) {
			errs = append(errs, prefix+".contract_digest must be sha256")
		}
		if seen[approval.ContractDigest] {
			errs = append(errs, prefix+" duplicates contract digest")
		}
		seen[approval.ContractDigest] = true
		if strings.TrimSpace(approval.ApprovedBy) == "" || strings.TrimSpace(approval.Justification) == "" {
			errs = append(errs, prefix+" approved_by and justification must not be blank")
		}
		if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(approval.ApprovedAt) {
			errs = append(errs, prefix+".approved_at must use YYYY-MM-DD")
		}
	}
	return dedupeSorted(errs)
}

func LoadEvalDocuments(skillDir string, descriptor Descriptor) ([]EvalDocument, error) {
	dir, safe := safeArtifactPath(skillDir, descriptor.Evals.Directory)
	if !safe {
		return nil, fmt.Errorf("unsafe eval directory")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var documents []EvalDocument
	for _, entry := range entries {
		if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".yaml") && !strings.HasSuffix(entry.Name(), ".yml")) {
			continue
		}
		data, err := os.ReadFile(dir + string(os.PathSeparator) + entry.Name())
		if err != nil {
			return nil, err
		}
		document, err := ParseEval(data)
		if err != nil {
			return nil, err
		}
		documents = append(documents, document)
	}
	return documents, nil
}

func RequiredASPSProperties(document AssuranceDocument) []string {
	set := map[string]bool{}
	for _, profile := range document.ASPS.RequiredProfiles {
		for _, property := range asps.ProfileProperties(profile) {
			set[property] = true
		}
	}
	for _, property := range document.ASPS.RequiredProperties {
		set[property] = true
	}
	out := make([]string, 0, len(set))
	for property := range set {
		out = append(out, property)
	}
	sort.Strings(out)
	return out
}
