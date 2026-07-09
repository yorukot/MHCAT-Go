package utility_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUtilityFeatureAvoidsDiscordGoAndMongoDriver(t *testing.T) {
	forbidden := []string{
		"github.com/bwmarrin/" + "discord" + "go",
		"go.mongodb.org/" + "mongo" + "-driver",
	}
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		for _, value := range forbidden {
			if strings.Contains(content, value) {
				t.Fatalf("forbidden import %q found in %s", value, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk utility feature: %v", err)
	}
}
