package utility

import "testing"

func TestLegacyHelpMetadataCoversKnownCommandInventory(t *testing.T) {
	commands := make(map[string]struct{})
	for _, category := range legacyHelpCategories {
		for _, command := range category.Commands {
			if _, duplicate := commands[command.Name]; duplicate {
				t.Fatalf("duplicate help command %q", command.Name)
			}
			commands[command.Name] = struct{}{}
		}
	}
	if len(commands) != 74 {
		t.Fatalf("help commands = %d, want 74", len(commands))
	}
	if len(legacyHelpUserPerms) != 50 {
		t.Fatalf("permission metadata = %d, want 50", len(legacyHelpUserPerms))
	}
	if len(legacyHelpVideos) != 35 {
		t.Fatalf("video metadata = %d, want 35", len(legacyHelpVideos))
	}
	for name := range legacyHelpUserPerms {
		if _, exists := commands[name]; !exists {
			t.Errorf("permission metadata references unknown command %q", name)
		}
	}
	for name := range legacyHelpVideos {
		if _, exists := commands[name]; !exists {
			t.Errorf("video metadata references unknown command %q", name)
		}
	}
}
