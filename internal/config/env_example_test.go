package config_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestEnvExampleListsRuntimeEnvKeys(t *testing.T) {
	repoRoot := filepath.Clean("../..")
	exampleData, err := os.ReadFile(filepath.Join(repoRoot, ".env.example"))
	if err != nil {
		t.Fatalf("read .env.example: %v", err)
	}
	exampleKeys := map[string]struct{}{}
	for _, line := range strings.Split(string(exampleData), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, _, ok := strings.Cut(line, "=")
		if ok && strings.HasPrefix(key, "MHCAT_") {
			exampleKeys[key] = struct{}{}
		}
	}

	codeKeys := map[string]struct{}{}
	envKeyRe := regexp.MustCompile(`"(MHCAT_[A-Z0-9_]+)"`)
	for _, root := range []string{
		filepath.Join(repoRoot, "internal", "config"),
		filepath.Join(repoRoot, "cmd"),
	} {
		if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for _, match := range envKeyRe.FindAllStringSubmatch(string(data), -1) {
				codeKeys[match[1]] = struct{}{}
			}
			return nil
		}); err != nil {
			t.Fatalf("scan %s: %v", root, err)
		}
	}

	missing := []string{}
	for key := range codeKeys {
		if _, ok := exampleKeys[key]; !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf(".env.example is missing runtime env keys: %s", strings.Join(missing, ", "))
	}
}
