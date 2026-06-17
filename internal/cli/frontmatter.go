package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// validateSkillArg rejects skill name arguments that contain path separators or
// ".." components, preventing traversal outside the skill source directory.
func validateSkillArg(name string) error {
	if filepath.Base(name) != name || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("invalid skill name %q: must be a plain directory name", name)
	}
	return nil
}

// parseFrontmatterScalar returns the string value of a top-level scalar key
// in YAML frontmatter by unmarshalling into a map. Returns "" if not found.
func parseFrontmatterScalar(content, key string) string {
	fm := parseFrontmatterMap(content)
	if fm == nil {
		return ""
	}
	v, ok := fm[key]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// parseFrontmatterSlice returns the string elements of a top-level sequence key
// in YAML frontmatter. Returns nil if the key is absent or not a sequence.
func parseFrontmatterSlice(content, key string) []string {
	fm := parseFrontmatterMap(content)
	if fm == nil {
		return nil
	}
	v, ok := fm[key]
	if !ok {
		return nil
	}
	// Handle both []interface{} from yaml.v3 and []string direct.
	switch val := v.(type) {
	case []interface{}:
		out := make([]string, 0, len(val))
		for _, item := range val {
			out = append(out, fmt.Sprintf("%v", item))
		}
		return out
	case []string:
		return val
	case string:
		if val == "" {
			return nil
		}
		return []string{val}
	}
	return nil
}

// parseFrontmatterMap extracts and unmarshals the YAML frontmatter block.
// Returns nil if the content does not have a valid frontmatter fence.
func parseFrontmatterMap(content string) map[string]any {
	if !strings.HasPrefix(content, "---\n") {
		return nil
	}
	end := strings.Index(content[4:], "\n---")
	if end < 0 {
		return nil
	}
	var fm map[string]any
	if err := yaml.Unmarshal([]byte(content[4:4+end]), &fm); err != nil {
		return nil
	}
	return fm
}
