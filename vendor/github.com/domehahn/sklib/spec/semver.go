package spec

import (
	"fmt"
	"strconv"
	"strings"
)

// IsSemVer returns true if value is a valid SemVer string (stable or pre-release/build).
// Leading "v" is accepted.
func IsSemVer(value string) bool {
	_, err := parseSemVerParts(value)
	return err == nil
}

// IsStableSemVer returns true if value is a stable SemVer (MAJOR.MINOR.PATCH, no pre-release or build metadata).
// Leading "v" is accepted.
func IsStableSemVer(value string) bool {
	parts, err := parseSemVerParts(value)
	if err != nil {
		return false
	}
	return parts.preRelease == "" && parts.buildMeta == ""
}

// NormalizeVersion strips a leading "v" and validates the result as SemVer.
// Returns an error for empty input, leading "v" with nothing after, or malformed SemVer.
func NormalizeVersion(value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("version must not be empty")
	}
	stripped := strings.TrimPrefix(value, "v")
	if stripped == "" {
		return "", fmt.Errorf("version %q is invalid: nothing after leading 'v'", value)
	}
	if _, err := parseSemVerParts(stripped); err != nil {
		return "", fmt.Errorf("version %q is not valid SemVer: %w", value, err)
	}
	return stripped, nil
}

// CompareVersionsSafe compares two SemVer strings.
// Returns -1 if a < b, 0 if a == b, 1 if a > b, and an error if either value is not valid SemVer.
func CompareVersionsSafe(a, b string) (int, error) {
	pa, err := parseSemVerParts(a)
	if err != nil {
		return 0, fmt.Errorf("invalid version %q: %w", a, err)
	}
	pb, err := parseSemVerParts(b)
	if err != nil {
		return 0, fmt.Errorf("invalid version %q: %w", b, err)
	}
	for _, pair := range [][2]int{{pa.major, pb.major}, {pa.minor, pb.minor}, {pa.patch, pb.patch}} {
		if pair[0] < pair[1] {
			return -1, nil
		}
		if pair[0] > pair[1] {
			return 1, nil
		}
	}
	return 0, nil
}

// CompareVersions compares two SemVer strings (stable only; pre-release ordering is not guaranteed).
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
// Panics if either value is not valid SemVer.
//
// Deprecated: use CompareVersionsSafe in all production code. CompareVersions is kept only as a
// sort.Slice convenience where both inputs are guaranteed pre-validated.
func CompareVersions(a, b string) int {
	pa, err := parseSemVerParts(a)
	if err != nil {
		panic(fmt.Sprintf("CompareVersions: invalid version %q: %v", a, err))
	}
	pb, err := parseSemVerParts(b)
	if err != nil {
		panic(fmt.Sprintf("CompareVersions: invalid version %q: %v", b, err))
	}

	for _, pair := range [][2]int{{pa.major, pb.major}, {pa.minor, pb.minor}, {pa.patch, pb.patch}} {
		if pair[0] < pair[1] {
			return -1
		}
		if pair[0] > pair[1] {
			return 1
		}
	}
	return 0
}

type semVerParts struct {
	major      int
	minor      int
	patch      int
	preRelease string
	buildMeta  string
}

func parseSemVerParts(value string) (semVerParts, error) {
	v := strings.TrimPrefix(value, "v")
	if v == "" {
		return semVerParts{}, fmt.Errorf("empty version string")
	}

	// Split off build metadata
	var buildMeta string
	if idx := strings.Index(v, "+"); idx >= 0 {
		buildMeta = v[idx+1:]
		v = v[:idx]
	}

	// Split off pre-release
	var preRelease string
	if idx := strings.Index(v, "-"); idx >= 0 {
		preRelease = v[idx+1:]
		v = v[:idx]
	}

	segments := strings.Split(v, ".")
	if len(segments) != 3 {
		return semVerParts{}, fmt.Errorf("expected MAJOR.MINOR.PATCH, got %q", value)
	}

	nums := make([]int, 3)
	for i, s := range segments {
		if s == "" {
			return semVerParts{}, fmt.Errorf("empty version segment in %q", value)
		}
		if len(s) > 1 && s[0] == '0' {
			return semVerParts{}, fmt.Errorf("leading zero in version segment %q", s)
		}
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			return semVerParts{}, fmt.Errorf("non-numeric version segment %q in %q", s, value)
		}
		nums[i] = n
	}

	return semVerParts{
		major:      nums[0],
		minor:      nums[1],
		patch:      nums[2],
		preRelease: preRelease,
		buildMeta:  buildMeta,
	}, nil
}
