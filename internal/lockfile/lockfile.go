package lockfile

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/agentic-template-kit/skcr/internal/models"
	"gopkg.in/yaml.v3"
)

const LockfileName = ".agentic-template.lock"

func Sha256Text(value string) string {
	hash := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(hash[:])
}

func WriteLockfile(targetDir string, files []models.RenderedFile, targetName string) error {
	managed := make([]map[string]any, 0, len(files))
	for _, rendered := range files {
		managed = append(managed, map[string]any{
			"path":     rendered.Destination,
			"platform": rendered.Platform,
			"source":   rendered.Source,
			"checksum": Sha256Text(rendered.Content),
		})
	}
	payload := map[string]any{
		"version":       "1",
		"target":        targetName,
		"managed_files": managed,
	}
	encoded, err := yaml.Marshal(payload)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(targetDir, LockfileName), encoded, 0o644)
}

func LoadLockfile(targetDir string) (map[string]any, error) {
	path := filepath.Join(targetDir, LockfileName)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoded := map[string]any{}
	if err := yaml.Unmarshal(payload, &decoded); err != nil {
		return nil, err
	}
	if decoded == nil {
		return map[string]any{}, nil
	}
	return decoded, nil
}

func ManagedFilesByPath(lock map[string]any) map[string]map[string]any {
	result := map[string]map[string]any{}
	raw, ok := lock["managed_files"].([]any)
	if !ok {
		return result
	}
	for _, item := range raw {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		path, _ := entry["path"].(string)
		if path == "" {
			continue
		}
		result[path] = entry
	}
	return result
}
