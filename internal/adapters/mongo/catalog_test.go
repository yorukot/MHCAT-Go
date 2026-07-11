package mongo

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestDefaultCollectionCatalogValid(t *testing.T) {
	catalog := DefaultCollectionCatalog()
	if err := ValidateCollectionCatalog(catalog); err != nil {
		t.Fatalf("validate catalog: %v", err)
	}
	if len(catalog) != 47 {
		t.Fatalf("catalog length = %d, want 47", len(catalog))
	}
}

func TestDefaultCollectionCatalogCoversLegacyModelFiles(t *testing.T) {
	legacyFiles := legacyModelFiles(t)
	byFile := CollectionCatalogByLegacyFile(DefaultCollectionCatalog())
	if len(byFile) != len(legacyFiles) {
		t.Fatalf("catalog file count = %d, legacy file count = %d", len(byFile), len(legacyFiles))
	}
	for _, file := range legacyFiles {
		if _, ok := byFile[file]; !ok {
			t.Fatalf("catalog missing legacy model file %s", file)
		}
	}
	for file := range byFile {
		if !containsString(legacyFiles, file) {
			t.Fatalf("catalog references missing legacy model file %s", file)
		}
	}
}

func TestDefaultCollectionCatalogUsesMongooseCollectionNames(t *testing.T) {
	byModel := CollectionCatalogByModel(DefaultCollectionCatalog())
	cases := map[string]string{
		"all_use_count":    "all_use_counts",
		"coin":             "coins",
		"guild":            "guilds",
		"join_message":     "join_messages",
		"message_reaction": "message_reactions",
		"role_number":      "role_numbers",
		"text_xp":          "text_xps",
		"voice_xp":         "voice_xps",
		"work_something":   "work_somethings",
	}
	for model, wantCollection := range cases {
		spec, ok := byModel[model]
		if !ok {
			t.Fatalf("model %s not found", model)
		}
		if spec.Name != wantCollection {
			t.Fatalf("model %s collection = %s, want %s", model, spec.Name, wantCollection)
		}
	}
}

func TestDefaultCollectionCatalogUniqueIndexesRequireDuplicateAudit(t *testing.T) {
	for _, spec := range DefaultCollectionCatalog() {
		for _, index := range spec.PlannedIndexes {
			if index.Unique && !index.RequiresDuplicateAudit {
				t.Fatalf("unique index %s.%s does not require duplicate audit", spec.Name, index.Name)
			}
		}
	}
}

func TestRoleSelectionIndexesRemainExplicitlyAuditGated(t *testing.T) {
	byName := CollectionCatalogByName(DefaultCollectionCatalog())
	tests := []struct {
		collection string
		index      string
	}{
		{collection: "btns", index: "btns_guild_number"},
		{collection: "message_reactions", index: "message_reactions_guild_message_react"},
	}
	for _, test := range tests {
		spec := byName[test.collection]
		var index *IndexSpec
		for candidateIndex := range spec.PlannedIndexes {
			candidate := &spec.PlannedIndexes[candidateIndex]
			if candidate.Name == test.index {
				index = candidate
				break
			}
		}
		if index == nil || !index.Unique || !index.RequiresDuplicateAudit {
			t.Fatalf("%s index = %#v", test.collection, index)
		}
	}
}

func TestDefaultCollectionCatalogLookupMaps(t *testing.T) {
	catalog := DefaultCollectionCatalog()
	byName := CollectionCatalogByName(catalog)
	byModel := CollectionCatalogByModel(catalog)
	byFile := CollectionCatalogByLegacyFile(catalog)
	if byName["coins"].LegacyMongooseModel != "coin" {
		t.Fatalf("coins lookup = %#v", byName["coins"])
	}
	if byModel["role_number"].LegacyModelFile != "models/role.js" {
		t.Fatalf("role_number lookup = %#v", byModel["role_number"])
	}
	if byFile["models/Number.js"].Name != "numbers" {
		t.Fatalf("Number.js lookup = %#v", byFile["models/Number.js"])
	}
}

func TestDefaultIndexPlanUsesCorrectedCollections(t *testing.T) {
	plan := DefaultIndexPlan(DefaultCollectionCatalog())
	if err := plan.Validate(); err != nil {
		t.Fatalf("validate default index plan: %v", err)
	}
	for _, index := range plan.Indexes {
		switch index.Collection {
		case "coin", "text_xp", "voice_xp", "poll", "ticket", "guild", "cron_set", "verification", "chatgpt":
			t.Fatalf("default index plan still uses singular scaffold collection %s", index.Collection)
		}
	}
}

func TestPerformanceIndexesDoNotRequireDuplicateCleanup(t *testing.T) {
	byName := CollectionCatalogByName(DefaultCollectionCatalog())
	for _, test := range []struct {
		collection string
		index      string
	}{
		{collection: "ann_all_sets", index: "ann_all_sets_guild_announcement_lookup"},
		{collection: "birthday_sets", index: "birthday_sets_guild_lookup"},
		{collection: "birthdays", index: "birthdays_guild_user_lookup"},
		{collection: "chats", index: "chats_guild_lookup"},
		{collection: "coins", index: "coins_guild_member_lookup"},
		{collection: "coins", index: "coins_today_guild"},
		{collection: "gifts", index: "gifts_guild_gift_name_lookup"},
		{collection: "gift_changes", index: "gift_changes_guild_lookup"},
		{collection: "ghps", index: "ghps_guild_commodity_id_lookup"},
		{collection: "join_messages", index: "join_messages_guild_lookup"},
		{collection: "join_roles", index: "join_roles_guild_lookup"},
		{collection: "leave_messages", index: "leave_messages_guild_lookup"},
		{collection: "loggings", index: "loggings_guild_lookup"},
		{collection: "message_reactions", index: "message_reactions_guild_message_react_lookup"},
		{collection: "polls", index: "polls_guild_messageid_lookup"},
		{collection: "text_xps", index: "text_xps_guild_member_lookup"},
		{collection: "voice_xps", index: "voice_xps_guild_member_lookup"},
		{collection: "voice_channels", index: "voice_channels_guild_parent_lookup"},
		{collection: "work_users", index: "work_users_guild_energi"},
		{collection: "good_webs", index: "good_webs_guild_lookup"},
	} {
		var found *IndexSpec
		for i := range byName[test.collection].PlannedIndexes {
			candidate := &byName[test.collection].PlannedIndexes[i]
			if candidate.Name == test.index {
				found = candidate
				break
			}
		}
		if found == nil || found.Unique || found.RequiresDuplicateAudit {
			t.Fatalf("%s index = %#v", test.index, found)
		}
	}
}

func legacyModelFiles(t *testing.T) []string {
	t.Helper()
	dir := filepath.Join("..", "..", "..", "..", "MHCAT", "models")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read legacy model directory %s: %v", dir, err)
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".js" {
			continue
		}
		files = append(files, filepath.ToSlash(filepath.Join("models", entry.Name())))
	}
	sort.Strings(files)
	return files
}

func containsString(values []string, value string) bool {
	i := sort.SearchStrings(values, value)
	return i < len(values) && values[i] == value
}
