// Package cncf exposes the official CNCF Landscape snapshot as deterministic
// skcr skill metadata.
package cncf

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"sync"
	"unicode"

	"gopkg.in/yaml.v3"
)

const SourceURL = "https://raw.githubusercontent.com/cncf/landscape/master/landscape.yml"

//go:generate go run ./generate -out landscape.yml

//go:embed landscape.yml
var landscapeYAML []byte

type Placement struct {
	Category    string
	Subcategory string
}

type Entry struct {
	Name        string
	SkillName   string
	Description string
	HomepageURL string
	RepoURL     string
	Project     string
	Placements  []Placement
}

type landscapeDocument struct {
	Landscape []landscapeCategory `yaml:"landscape"`
}

type landscapeCategory struct {
	Name          string                 `yaml:"name"`
	Subcategories []landscapeSubcategory `yaml:"subcategories"`
}

type landscapeSubcategory struct {
	Name  string          `yaml:"name"`
	Items []landscapeItem `yaml:"items"`
}

type landscapeItem struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	HomepageURL string `yaml:"homepage_url"`
	RepoURL     string `yaml:"repo_url"`
	Project     string `yaml:"project"`
}

var (
	loadOnce sync.Once
	entries  []Entry
	loadErr  error
)

func Entries() ([]Entry, error) {
	loadOnce.Do(func() {
		entries, loadErr = parseLandscape(landscapeYAML)
	})
	return cloneEntries(entries), loadErr
}

func MustEntries() []Entry {
	items, err := Entries()
	if err != nil {
		panic(fmt.Sprintf("parse embedded CNCF Landscape: %v", err))
	}
	return items
}

func SkillNames() []string {
	items := MustEntries()
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.SkillName)
	}
	return names
}

// SkillCategories returns active CNCF projects, the complete Landscape,
// maturity groups, semantic technology facets, and every official top-level
// category and subcategory. A product appearing more than once remains one
// skill and is included in every corresponding category.
func SkillCategories() map[string][]string {
	result := map[string][]string{"cncf": {}, "cncf-landscape": {}}
	seen := map[string]map[string]bool{"cncf": {}, "cncf-landscape": {}}
	for _, item := range MustEntries() {
		addCategorySkill(result, seen, "cncf-landscape", item.SkillName)
		switch item.Project {
		case "graduated", "incubating", "sandbox":
			addCategorySkill(result, seen, "cncf", item.SkillName)
			addCategorySkill(result, seen, "cncf-"+item.Project, item.SkillName)
		case "archived":
			addCategorySkill(result, seen, "cncf-archived", item.SkillName)
		}
		for _, placement := range item.Placements {
			top := categoryName("cncf", placement.Category)
			leaf := categoryName(top, placement.Subcategory)
			addCategorySkill(result, seen, top, item.SkillName)
			addCategorySkill(result, seen, leaf, item.SkillName)
			for _, category := range semanticCategories(placement) {
				addCategorySkill(result, seen, category, item.SkillName)
			}
		}
	}
	for name := range result {
		sort.Strings(result[name])
	}
	return result
}

func SnapshotDigest() string {
	sum := sha256.Sum256(landscapeYAML)
	return fmt.Sprintf("sha256:%x", sum[:])
}

func parseLandscape(payload []byte) ([]Entry, error) {
	var document landscapeDocument
	if err := yaml.Unmarshal(payload, &document); err != nil {
		return nil, err
	}
	byName := map[string]*Entry{}
	for _, category := range document.Landscape {
		for _, subcategory := range category.Subcategories {
			for _, item := range subcategory.Items {
				name := strings.TrimSpace(item.Name)
				if name == "" {
					continue
				}
				entry := byName[name]
				if entry == nil {
					entry = &Entry{Name: name, Description: strings.TrimSpace(item.Description), HomepageURL: item.HomepageURL, RepoURL: item.RepoURL, Project: item.Project}
					byName[name] = entry
				} else {
					fillEntryMetadata(entry, item)
				}
				placement := Placement{Category: category.Name, Subcategory: subcategory.Name}
				if !containsPlacement(entry.Placements, placement) {
					entry.Placements = append(entry.Placements, placement)
				}
			}
		}
	}

	items := make([]Entry, 0, len(byName))
	baseOwners := map[string][]string{}
	for name := range byName {
		base := slug(name)
		baseOwners[base] = append(baseOwners[base], name)
	}
	for name, entry := range byName {
		base := slug(name)
		if len(baseOwners[base]) > 1 {
			base = withHash(base, shortHash(name), 49)
		}
		entry.SkillName = "cncf-" + boundedSlug(base, shortHash(name), 49) + "-reviewer"
		sort.Slice(entry.Placements, func(i, j int) bool {
			left := entry.Placements[i].Category + "\x00" + entry.Placements[i].Subcategory
			right := entry.Placements[j].Category + "\x00" + entry.Placements[j].Subcategory
			return left < right
		})
		items = append(items, *entry)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].SkillName < items[j].SkillName })
	return items, nil
}

func fillEntryMetadata(entry *Entry, item landscapeItem) {
	if entry.Description == "" {
		entry.Description = strings.TrimSpace(item.Description)
	}
	if entry.HomepageURL == "" {
		entry.HomepageURL = item.HomepageURL
	}
	if entry.RepoURL == "" {
		entry.RepoURL = item.RepoURL
	}
	if entry.Project == "" {
		entry.Project = item.Project
	}
}

func containsPlacement(items []Placement, target Placement) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func cloneEntries(source []Entry) []Entry {
	result := make([]Entry, len(source))
	for i, item := range source {
		result[i] = item
		result[i].Placements = append([]Placement(nil), item.Placements...)
	}
	return result
}

func addCategorySkill(result map[string][]string, seen map[string]map[string]bool, category, skill string) {
	if seen[category] == nil {
		seen[category] = map[string]bool{}
	}
	if seen[category][skill] {
		return
	}
	seen[category][skill] = true
	result[category] = append(result[category], skill)
}

func categoryName(prefix, value string) string {
	part := slug(value)
	base := strings.Trim(prefix+"-"+part, "-")
	if prefix == "cncf" && strings.HasPrefix(part, "cncf-") {
		base = part
	}
	if len(base) <= 64 {
		return base
	}
	return boundedSlug(base, shortHash(base), 64)
}

func slug(value string) string {
	value = strings.NewReplacer("+", " plus ", "#", " sharp ", "&", " and ", "@", " at ").Replace(strings.ToLower(value))
	var result strings.Builder
	separator := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			if separator && result.Len() > 0 {
				result.WriteByte('-')
			}
			result.WriteRune(r)
			separator = false
		case unicode.IsLetter(r), unicode.IsDigit(r):
			separator = true
		default:
			separator = true
		}
	}
	clean := strings.Trim(result.String(), "-")
	if clean == "" {
		return "item-" + shortHash(value)
	}
	return clean
}

func boundedSlug(value, hash string, limit int) string {
	value = strings.Trim(value, "-")
	if len(value) <= limit {
		return value
	}
	prefixLength := limit - len(hash) - 1
	if prefixLength < 1 {
		return hash[:limit]
	}
	prefix := strings.Trim(value[:prefixLength], "-")
	return prefix + "-" + hash
}

func withHash(value, hash string, limit int) string {
	prefixLength := limit - len(hash) - 1
	if prefixLength < 1 {
		return hash[:limit]
	}
	if len(value) > prefixLength {
		value = value[:prefixLength]
	}
	return strings.Trim(value, "-") + "-" + hash
}

func shortHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", sum[:4])
}
