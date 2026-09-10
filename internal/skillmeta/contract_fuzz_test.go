package skillmeta_test

import (
	"testing"

	"github.com/domehahn/skcr/v2/internal/skillmeta"
)

func FuzzParseDescriptor(f *testing.F) {
	f.Add([]byte("schema_version: \"2.0.0\"\nname: fuzz-test\nversion: \"1.0.0\"\ndescription: fuzzing descriptor\ncontract: {file: contract.yaml}\n"))
	f.Add([]byte("invalid descriptor payload"))

	f.Fuzz(func(t *testing.T, payload []byte) {
		_, _ = skillmeta.ParseDescriptor(payload)
	})
}

func FuzzParseContract(f *testing.F) {
	f.Add([]byte("schema_version: \"2.0.0\"\ncapabilities:\n  runtime:\n    required:\n      filesystem: {read: [], write: [], delete: []}\n"))
	f.Add([]byte("invalid contract payload"))

	f.Fuzz(func(t *testing.T, payload []byte) {
		_, _ = skillmeta.ParseContract(payload)
	})
}

func FuzzParseEval(f *testing.F) {
	f.Add([]byte("schema_version: \"2.0.0\"\nscenarios:\n  - id: test-scenario\n    type: behavioral\n"))
	f.Add([]byte("invalid eval payload"))

	f.Fuzz(func(t *testing.T, payload []byte) {
		_, _ = skillmeta.ParseEval(payload)
	})
}
