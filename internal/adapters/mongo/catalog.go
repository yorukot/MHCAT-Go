package mongo

import (
	"fmt"
	"sort"
	"strings"
)

type catalogIndex struct {
	Name   string
	Fields []IndexKey
	Unique bool
	Reason string
}

func DefaultCollectionCatalog() []CollectionSpec {
	specs := []CollectionSpec{
		catalogSpec("numbers", "Number", "models/Number.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("numbers_guild", []string{"guild"}, "guild stats channel config singleton"),
		}, "Capitalized legacy model name; verify Mongoose pluralization against live collections."),
		catalogSpec("all_use_counts", "all_use_count", "models/all_use_count.js", []string{"slashcommand_name"}, []catalogIndex{
			uniqueCatalogIndex("all_use_counts_slashcommand_name", []string{"slashcommand_name"}, "slash command usage counter lookup"),
		}, "Usage writes are feature-gated; audit duplicate and null/blank command names before unique index apply."),
		catalogSpec("ann_all_sets", "ann_all_set", "models/ann_all_set.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("ann_all_sets_guild_announcement_id", []string{"guild", "announcement_id"}, "announcement config lookup by guild and announcement id"),
		}, ""),
		catalogSpec("birthdays", "birthday", "models/birthday.js", []string{"guild", "user"}, []catalogIndex{
			uniqueCatalogIndex("birthdays_guild_user", []string{"guild", "user"}, "user birthday lookup"),
		}, "Birthday send job is inactive in legacy; do not restore scheduler behavior without ADR."),
		catalogSpec("birthday_sets", "birthday_set", "models/birthday_set.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("birthday_sets_guild", []string{"guild"}, "birthday guild config singleton"),
		}, ""),
		catalogSpec("btns", "btn", "models/btn.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("btns_guild_number", []string{"guild", "number"}, "role button lookup"),
		}, "Go aligns duplicate role mappings during explicit setup and creates no startup index; apply uniqueness only after a full duplicate audit and exclusive Node/Go setup ownership."),
		catalogSpec("chats", "chat", "models/chat.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("chats_guild", []string{"guild"}, "autochat config singleton"),
		}, "Message Content intent and chat feature rollout remain gated."),
		catalogSpec("chat_roles", "chat_role", "models/chat_role.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("chat_roles_guild_leavel_role", []string{"guild", "leavel", "role"}, "text XP level role lookup"),
		}, "Preserve misspelled legacy field `leavel`."),
		catalogSpec("chatgpts", "chatgpt", "models/chatgpt.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("chatgpts_guild", []string{"guild"}, "ChatGPT handoff state lookup"),
		}, "Paid handoff writes exact legacy fields in a transaction; external worker and singleton ownership must be confirmed before enablement."),
		catalogSpec("chatgpt_gets", "chatgpt_get", "models/chatgpt_get.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("chatgpt_gets_guild", []string{"guild"}, "ChatGPT pricing/config lookup"),
		}, "Redeem credits and paid handoff debits share this collection; audit duplicates and worker ownership before writes."),
		catalogSpec("codes", "code", "models/code.js", []string{"code"}, []catalogIndex{
			uniqueCatalogIndex("codes_code", []string{"code"}, "redeem code lookup"),
		}, ""),
		catalogSpec("coins", "coin", "models/coin.js", []string{"guild", "member"}, []catalogIndex{
			uniqueCatalogIndex("coins_guild_member", []string{"guild", "member"}, "balance lookup and atomic coin updates"),
			nonUniqueCatalogIndex("coins_guild_coin_rank", []IndexKey{{Field: "guild", Order: 1}, {Field: "coin", Order: -1}}, "coin ranking"),
		}, "Economy writes require atomic repository methods; audit `today` mixed types before strict decode. Work payout adds rollback-compatible `mhcat_work_payouts` markers."),
		catalogSpec("create_hours", "create_hours", "models/create_hours.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("create_hours_guild", []string{"guild"}, "account-age join policy singleton"),
		}, "Mongoose pluralization for this underscore plural-looking model is intentionally preserved as `create_hours`."),
		catalogSpec("cron_sets", "cron_set", "models/cron_set.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("cron_sets_guild_id", []string{"guild", "id"}, "scheduled message lookup"),
		}, "Scheduler ownership/lease design is required before Go sends scheduled messages."),
		catalogSpec("errors_sets", "errors_set", "models/errors_set.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("errors_sets_guild", []string{"guild"}, "warning escalation config singleton"),
		}, ""),
		catalogSpec("ghps", "ghp", "models/ghp.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("ghps_guild_commodity_id", []string{"guild", "commodity_id"}, "shop item lookup"),
		}, "Legacy writes mix numeric/string values for shop prices and counts."),
		catalogSpec("gifts", "gift", "models/gift.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("gifts_guild_gift_name", []string{"guild", "gift_name"}, "gacha prize lookup"),
		}, "Preserve misspelled legacy field `gift_chence`."),
		catalogSpec("gift_changes", "gift_change", "models/gift_change.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("gift_changes_guild", []string{"guild"}, "economy/gacha config singleton"),
			nonUniqueCatalogIndex("gift_changes_time_guild", []IndexKey{{Field: "time", Order: 1}, {Field: "guild", Order: 1}}, "daily reset exclusion scan"),
		}, "Partial index semantics for `time != 0` need ADR before use."),
		catalogSpec("good_webs", "good_web", "models/good_web.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("good_webs_guild", []string{"guild"}, "anti-scam toggle singleton"),
		}, ""),
		catalogSpec("guilds", "guild", "models/guild.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("guilds_guild", []string{"guild"}, "guild config singleton"),
		}, "Dashboard-shared collection; patch writes only after compatibility review."),
		catalogSpec("join_messages", "join_message", "models/join_message.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("join_messages_guild", []string{"guild"}, "welcome message config singleton"),
		}, "Dashboard-shared collection; do not create legacy `enable` unique index."),
		catalogSpec("join_roles", "join_role", "models/join_role.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("join_roles_guild_role", []string{"guild", "role"}, "join role rule lookup"),
		}, ""),
		catalogSpec("leave_messages", "leave_message", "models/leave_message.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("leave_messages_guild", []string{"guild"}, "leave message config singleton"),
		}, ""),
		catalogSpec("lock_channels", "lock_channel", "models/lock_channel.js", []string{"guild", "channel_id"}, []catalogIndex{
			uniqueCatalogIndex("lock_channels_guild_channel_id", []string{"guild", "channel_id"}, "voice lock state lookup"),
		}, "Legacy stores lock answer text; avoid logging raw values."),
		catalogSpec("loggings", "logging", "models/logging.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("loggings_guild", []string{"guild"}, "logging config singleton"),
		}, "Audit-log reads require rate-limit policy before feature wiring."),
		catalogSpec("lotters", "lotter", "models/lotter.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("lotters_guild_id", []string{"guild", "id"}, "lottery lookup and member updates"),
		}, "Lottery creation is disabled in legacy; keep inactive behavior unless ADR changes it."),
		catalogSpec("message_reactions", "message_reaction", "models/message_reaction.js", []string{"guild", "message", "react"}, []catalogIndex{
			uniqueCatalogIndex("message_reactions_guild_message_react", []string{"guild", "message", "react"}, "reaction-role lookup"),
		}, "Dashboard backup evidence mentioned singular `message_reaction`; live audit must report it as unknown if present. Go aligns duplicates during setup and removes them on explicit delete; create no index until a full duplicate audit and exclusive Node/Go setup ownership."),
		catalogSpec("not_a_good_webs", "not_a_good_web", "models/not_a_good_web.js", []string{"web"}, []catalogIndex{
			uniqueCatalogIndex("not_a_good_webs_web", []string{"web"}, "anti-scam URL lookup"),
		}, "Normalize domains and escape regex before any schema/write changes."),
		catalogSpec("polls", "poll", "models/poll.js", []string{"guild", "messageid"}, []catalogIndex{
			uniqueCatalogIndex("polls_guild_messageid", []string{"guild", "messageid"}, "poll component lookup"),
		}, "Poll vote state requires atomic update design before writes."),
		catalogSpec("role_numbers", "role_number", "models/role.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("role_numbers_guild_role", []string{"guild", "role"}, "role stats config lookup"),
			uniqueCatalogIndex("role_numbers_guild_channel", []string{"guild", "channel"}, "role stats channel lookup"),
		}, "Legacy file is `role.js` but Mongoose model is `role_number`."),
		catalogSpec("sign_lists", "sign_list", "models/sign_list.js", []string{"guild", "member"}, []catalogIndex{
			uniqueCatalogIndex("sign_lists_guild_member", []string{"guild", "member"}, "sign-in history lookup"),
		}, ""),
		catalogSpec("suports", "suport", "models/suport.js", []string{"support_id"}, []catalogIndex{
			uniqueCatalogIndex("suports_support_id", []string{"support_id"}, "support lookup"),
		}, "Preserve misspelled legacy model/file name."),
		catalogSpec("systems", "system", "models/system.js", nil, nil, "Active usage unclear; no planned indexes until behavior is confirmed."),
		catalogSpec("text_xps", "text_xp", "models/text_xp.js", []string{"guild", "member"}, []catalogIndex{
			uniqueCatalogIndex("text_xps_guild_member", []string{"guild", "member"}, "text XP lookup and atomic XP increments"),
			nonUniqueCatalogIndex("text_xps_guild_rank", []IndexKey{{Field: "guild", Order: 1}, {Field: "leavel", Order: -1}, {Field: "xp", Order: -1}}, "text XP ranking after type audit"),
		}, "XP/level fields may be strings; audit before strict numeric sorting."),
		catalogSpec("text_xp_channels", "text_xp_channel", "models/text_xp_channel.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("text_xp_channels_guild", []string{"guild"}, "text XP announcement config singleton"),
		}, ""),
		catalogSpec("tickets", "ticket", "models/ticket.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("tickets_guild", []string{"guild"}, "ticket config singleton"),
		}, "Ticket feature must validate roles/channels before persisting config."),
		catalogSpec("verifications", "verification", "models/verification.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("verifications_guild", []string{"guild"}, "verification config singleton"),
		}, ""),
		catalogSpec("voice_channels", "voice_channel", "models/voice_channel.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("voice_channels_guild_ticket_channel", []string{"guild", "ticket_channel"}, "dynamic voice trigger config"),
		}, "Voice-state feature requires ownership/reconciliation before writes."),
		catalogSpec("voice_channel_ids", "voice_channel_id", "models/voice_channel_id.js", []string{"guild", "channel_id"}, []catalogIndex{
			uniqueCatalogIndex("voice_channel_ids_guild_channel_id", []string{"guild", "channel_id"}, "dynamic voice channel state lookup"),
		}, "Stale channel cleanup must be audit-first."),
		catalogSpec("voice_roles", "voice_role", "models/voice_role.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("voice_roles_guild_leavel_role", []string{"guild", "leavel", "role"}, "voice XP level role lookup"),
		}, "Preserve misspelled legacy field `leavel`."),
		catalogSpec("voice_xps", "voice_xp", "models/voice_xp.js", []string{"guild", "member"}, []catalogIndex{
			uniqueCatalogIndex("voice_xps_guild_member", []string{"guild", "member"}, "voice XP lookup and atomic XP increments"),
			nonUniqueCatalogIndex("voice_xps_guild_rank", []IndexKey{{Field: "guild", Order: 1}, {Field: "leavel", Order: -1}, {Field: "xp", Order: -1}}, "voice XP ranking after type audit"),
		}, "Voice session reconciliation is required before feature writes."),
		catalogSpec("voice_xp_channels", "voice_xp_channel", "models/voice_xp_channel.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("voice_xp_channels_guild", []string{"guild"}, "voice XP announcement config singleton"),
		}, ""),
		catalogSpec("votes", "vote", "models/vote.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("votes_guild_Number", []string{"guild", "Number"}, "vote state/config lookup"),
			uniqueCatalogIndex("votes_guild_member", []string{"guild", "member"}, "member vote lookup"),
		}, "Preserve capitalized legacy field `Number`; `member` is an array in legacy and needs shape audit."),
		catalogSpec("warndbs", "warndb", "models/warndb.js", []string{"guild", "user"}, []catalogIndex{
			nonUniqueCatalogIndex("warndbs_guild_user", []IndexKey{{Field: "guild", Order: 1}, {Field: "user", Order: 1}}, "warnings lookup without forcing one document per user"),
		}, "Dashboard-shared collection; multiple warning docs may be valid."),
		catalogSpec("work_sets", "work_set", "models/work_set.js", []string{"guild"}, []catalogIndex{
			uniqueCatalogIndex("work_sets_guild", []string{"guild"}, "work config singleton"),
		}, ""),
		catalogSpec("work_somethings", "work_something", "models/work_something.js", []string{"guild", "name"}, []catalogIndex{
			uniqueCatalogIndex("work_somethings_guild_name", []string{"guild", "name"}, "work task catalog lookup"),
		}, "Dashboard-shared collection; do not copy dashboard-only `guild` unique declaration."),
		catalogSpec("work_users", "work_user", "models/work_user.js", []string{"guild", "user"}, []catalogIndex{
			uniqueCatalogIndex("work_users_guild_user", []string{"guild", "user"}, "work user state lookup"),
			nonUniqueCatalogIndex("work_users_state_end_time", []IndexKey{{Field: "state", Order: 1}, {Field: "end_time", Order: 1}}, "due work-job scan"),
		}, "Preserve misspelled legacy field `energi`; work payout uses `work_users._id` for per-row idempotency."),
	}
	sortCollectionSpecs(specs)
	return specs
}

func ValidateCollectionCatalog(catalog []CollectionSpec) error {
	collections := map[string]struct{}{}
	models := map[string]struct{}{}
	files := map[string]struct{}{}
	for _, spec := range catalog {
		if strings.TrimSpace(spec.Name) == "" {
			return fmt.Errorf("catalog collection name is required")
		}
		if strings.TrimSpace(spec.LegacyMongooseModel) == "" {
			return fmt.Errorf("catalog %s legacy mongoose model is required", spec.Name)
		}
		if strings.TrimSpace(spec.LegacyModelFile) == "" {
			return fmt.Errorf("catalog %s legacy model file is required", spec.Name)
		}
		if _, ok := collections[spec.Name]; ok {
			return fmt.Errorf("duplicate catalog collection %s", spec.Name)
		}
		if _, ok := models[spec.LegacyMongooseModel]; ok {
			return fmt.Errorf("duplicate catalog mongoose model %s", spec.LegacyMongooseModel)
		}
		if _, ok := files[spec.LegacyModelFile]; ok {
			return fmt.Errorf("duplicate catalog legacy model file %s", spec.LegacyModelFile)
		}
		collections[spec.Name] = struct{}{}
		models[spec.LegacyMongooseModel] = struct{}{}
		files[spec.LegacyModelFile] = struct{}{}
		for _, index := range spec.PlannedIndexes {
			if index.Collection != spec.Name {
				return fmt.Errorf("index %s collection %s does not match catalog collection %s", index.Name, index.Collection, spec.Name)
			}
			if index.Unique && !index.RequiresDuplicateAudit {
				return fmt.Errorf("unique index %s.%s must require duplicate audit", spec.Name, index.Name)
			}
			if index.TTLSeconds != nil && !index.RequiresRetentionADR {
				return fmt.Errorf("ttl index %s.%s must require retention adr", spec.Name, index.Name)
			}
		}
	}
	return nil
}

func CollectionCatalogByName(catalog []CollectionSpec) map[string]CollectionSpec {
	result := make(map[string]CollectionSpec, len(catalog))
	for _, spec := range catalog {
		result[spec.Name] = spec
	}
	return result
}

func CollectionCatalogByModel(catalog []CollectionSpec) map[string]CollectionSpec {
	result := make(map[string]CollectionSpec, len(catalog))
	for _, spec := range catalog {
		result[spec.LegacyMongooseModel] = spec
	}
	return result
}

func CollectionCatalogByLegacyFile(catalog []CollectionSpec) map[string]CollectionSpec {
	result := make(map[string]CollectionSpec, len(catalog))
	for _, spec := range catalog {
		result[spec.LegacyModelFile] = spec
	}
	return result
}

func catalogSpec(collection, model, file string, required []string, indexes []catalogIndex, notes string) CollectionSpec {
	requiredFields := make([]FieldSpec, 0, len(required))
	for _, field := range required {
		requiredFields = append(requiredFields, FieldSpec{Name: field, Type: "string", Required: true})
	}
	logicalKeys := make([]LogicalKeySpec, 0, len(indexes))
	plannedIndexes := make([]IndexSpec, 0, len(indexes))
	for _, index := range indexes {
		fields := indexFieldNames(index.Fields)
		logicalKeys = append(logicalKeys, LogicalKeySpec{Name: index.Name, Fields: fields, Unique: index.Unique})
		planned := IndexSpec{
			Collection: collection,
			Name:       index.Name,
			Keys:       append([]IndexKey(nil), index.Fields...),
			Unique:     index.Unique,
			Reason:     index.Reason,
		}
		if index.Unique {
			planned.RequiresDuplicateAudit = true
		}
		plannedIndexes = append(plannedIndexes, planned)
	}
	return CollectionSpec{
		Name:                collection,
		LegacyModelFile:     file,
		LegacyMongooseModel: model,
		RequiredFields:      requiredFields,
		LogicalKeys:         logicalKeys,
		PlannedIndexes:      plannedIndexes,
		Incomplete:          true,
		Notes:               notesWithDefault(notes),
	}
}

func uniqueCatalogIndex(name string, fields []string, reason string) catalogIndex {
	return catalogIndex{Name: name, Fields: keysFromFields(fields), Unique: true, Reason: reason}
}

func nonUniqueCatalogIndex(name string, fields []IndexKey, reason string) catalogIndex {
	return catalogIndex{Name: name, Fields: append([]IndexKey(nil), fields...), Reason: reason}
}

func keysFromFields(fields []string) []IndexKey {
	keys := make([]IndexKey, 0, len(fields))
	for _, field := range fields {
		keys = append(keys, IndexKey{Field: field, Order: 1})
	}
	return keys
}

func indexFieldNames(keys []IndexKey) []string {
	fields := make([]string, 0, len(keys))
	for _, key := range keys {
		fields = append(fields, key.Field)
	}
	return fields
}

func notesWithDefault(notes string) string {
	const suffix = "Schema field coverage remains incomplete until BSON fixtures and live audit are reconciled; do not use this catalog to justify writes without feature-level tests."
	if strings.TrimSpace(notes) == "" {
		return suffix
	}
	return notes + " " + suffix
}

func sortCollectionSpecs(specs []CollectionSpec) {
	sort.SliceStable(specs, func(i, j int) bool {
		return specs[i].Name < specs[j].Name
	})
}
