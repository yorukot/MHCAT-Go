package customid

import (
	"regexp"
	"strconv"
	"strings"
)

type legacyRule struct {
	Name    string
	Pattern string
	Feature string
	Action  string
	Broad   bool
}

var (
	legacySnowflakeRe    = regexp.MustCompile(`^[0-9]{17,20}$`)
	legacySignRe         = regexp.MustCompile(`^/([0-9]{17,20})_sing\{([0-9]{4})\}-\[([0-9]{1,2})\]$`)
	legacyProfileRe      = regexp.MustCompile(`^([0-9]{17,20})my-profile$`)
	legacyRankRe         = regexp.MustCompile(`^\[([0-9]{17,20})\]\{([0-9]+)\}(text_rank|voice_rank|coin_rank)$`)
	legacyRankAltRe      = regexp.MustCompile(`^\[([0-9]{17,20})\](text_rank|voice_rank|coin_rank) \{([0-9]+)\}$`)
	legacyLeaveRoleRe    = regexp.MustCompile(`^([0-9]+)(text_leave_role|voice_leave_role)$`)
	legacyVerificationRe = regexp.MustCompile(`^[A-Za-z0-9]{1,16}verification$`)
	legacyPollRe         = regexp.MustCompile(`^poll_[^\x00-\x1F:]{1,80}$`)
	legacyRoleButtonRe   = regexp.MustCompile(`^[0-9]{12,32}(add|delete)$`)
	legacyRoleAddModalRe = regexp.MustCompile(`^roleaddcontent[0-9]{12,32}$`)
	legacyLotteryRe      = regexp.MustCompile(`^[A-Za-z0-9_-]{1,40}(search|restart|stop)?$`)
	legacyShopItemRe     = regexp.MustCompile(`^[A-Za-z0-9_-]{1,40}$`)
	legacyShopDetailRe   = regexp.MustCompile(`^[A-Za-z0-9_-]{1,40}ghp$`)
	legacyShopNumberRe   = regexp.MustCompile(`^([0-9]|back|confirm)ghp_number[A-Za-z0-9_-]*$`)
)

var exactComponentRoutes = map[string]RouteKey{
	"helphelphelphelpmenu": {Kind: InteractionKindComponent, Feature: "help", Action: "category_select", Version: LegacyVersion, Legacy: true},
	"help-menus":           {Kind: InteractionKindComponent, Feature: "help", Action: "legacy_unused", Version: LegacyVersion, Legacy: true},
	"tic":                  {Kind: InteractionKindComponent, Feature: "ticket", Action: "open", Version: LegacyVersion, Legacy: true},
	"del":                  {Kind: InteractionKindComponent, Feature: "ticket", Action: "close", Version: LegacyVersion, Legacy: true},
	"see_result":           {Kind: InteractionKindComponent, Feature: "poll", Action: "result", Version: LegacyVersion, Legacy: true},
	"poll_menu":            {Kind: InteractionKindComponent, Feature: "poll", Action: "owner_menu", Version: LegacyVersion, Legacy: true},
	"menu_choose":          {Kind: InteractionKindComponent, Feature: "poll", Action: "max_choices", Version: LegacyVersion, Legacy: true},
	"delete-data":          {Kind: InteractionKindComponent, Feature: "admin", Action: "delete_data_select", Version: LegacyVersion, Legacy: true},
	"loggin_create":        {Kind: InteractionKindComponent, Feature: "logging", Action: "configure_select", Version: LegacyVersion, Legacy: true},
	"botinfoupdate":        {Kind: InteractionKindComponent, Feature: "info", Action: "bot_refresh", Version: LegacyVersion, Legacy: true},
	"shardinfoupdate":      {Kind: InteractionKindComponent, Feature: "info", Action: "shard_refresh", Version: LegacyVersion, Legacy: true},
	"lock_start":           {Kind: InteractionKindComponent, Feature: "voice_lock", Action: "prompt", Version: LegacyVersion, Legacy: true},
	"yesssss":              {Kind: InteractionKindComponent, Feature: "game", Action: "yes", Version: LegacyVersion, Legacy: true},
	"nooooo":               {Kind: InteractionKindComponent, Feature: "game", Action: "no", Version: LegacyVersion, Legacy: true},
	"main_no_card":         {Kind: InteractionKindComponent, Feature: "game", Action: "main_stand", Version: LegacyVersion, Legacy: true},
	"main_get_card":        {Kind: InteractionKindComponent, Feature: "game", Action: "main_hit", Version: LegacyVersion, Legacy: true},
	"user_no_card":         {Kind: InteractionKindComponent, Feature: "game", Action: "user_stand", Version: LegacyVersion, Legacy: true},
	"user_get_card":        {Kind: InteractionKindComponent, Feature: "game", Action: "user_hit", Version: LegacyVersion, Legacy: true},
	"lookmenumber":         {Kind: InteractionKindComponent, Feature: "game", Action: "show_number", Version: LegacyVersion, Legacy: true},
	"teach21point":         {Kind: InteractionKindComponent, Feature: "game", Action: "teach_21_point", Version: LegacyVersion, Legacy: true},
	"thansize":             {Kind: InteractionKindComponent, Feature: "game", Action: "than_size_help", Version: LegacyVersion, Legacy: true},
}

var ambiguousExactIDs = map[string]string{
	"announcement_yes": "announcement/work confirmation reuse",
	"announcement_no":  "announcement/work confirmation reuse",
	"week_menu":        "cron/birthday select reuse",
	"hour_menu":        "cron/birthday select reuse",
	"min_menu":         "cron/birthday select reuse",
}

var disabledRankIDs = map[string]struct{}{
	"text_rank": {}, "text_rank1": {}, "text_rank2": {}, "text_rank4": {}, "text_rank5": {},
	"coin_rank": {}, "coin_rank1": {}, "coin_rank2": {}, "coin_rank4": {}, "coin_rank5": {},
}

func ParseLegacyComponent(raw string) (ID, error) {
	if raw == "" {
		return ID{}, ErrEmptyID
	}
	if customIDLength(raw) > MaxCustomIDLength {
		return ID{}, safeError(ErrTooLong, "legacy component")
	}
	if reason, ok := ambiguousExactIDs[raw]; ok {
		return ID{}, safeError(ErrAmbiguousID, reason)
	}
	if key, ok := exactComponentRoutes[raw]; ok {
		return idFromRoute(raw, key), nil
	}
	if _, ok := disabledRankIDs[raw]; ok {
		return idFromRoute(raw, RouteKey{Kind: InteractionKindComponent, Feature: "rank", Action: "disabled_legacy", Version: LegacyVersion, Legacy: true}), nil
	}
	if id, ok, err := parseRank(raw); ok || err != nil {
		return id, err
	}
	if id, ok, err := parseSign(raw); ok || err != nil {
		return id, err
	}
	if legacyProfileRe.MatchString(raw) {
		return legacyComponent("profile", "refresh", raw), nil
	}
	if id, ok, err := parseLeaveRole(raw); ok || err != nil {
		return id, err
	}
	if legacyPollRe.MatchString(raw) {
		return legacyComponent("poll", "vote", raw), nil
	}
	if legacyVerificationRe.MatchString(raw) {
		return legacyComponent("verification", "prompt", raw), nil
	}
	if legacyRoleButtonRe.MatchString(raw) {
		if strings.HasSuffix(raw, "add") {
			return legacyComponent("role_button", "add", raw), nil
		}
		return legacyComponent("role_button", "remove", raw), nil
	}
	if legacyShopNumberRe.MatchString(raw) {
		return legacyComponent("shop", "quantity", raw), nil
	}
	if legacyShopDetailRe.MatchString(raw) && !strings.HasSuffix(raw, "ghp_number") {
		return legacyComponent("shop", "detail", raw), nil
	}
	if routeLottery(raw) {
		switch {
		case strings.HasSuffix(raw, "search"):
			return legacyComponent("lottery", "search", raw), nil
		case strings.HasSuffix(raw, "restart"):
			return legacyComponent("lottery", "reroll", raw), nil
		case strings.HasSuffix(raw, "stop"):
			return legacyComponent("lottery", "stop", raw), nil
		}
		return legacyComponent("lottery", "enter", raw), nil
	}
	if legacyShopItemRe.MatchString(raw) {
		return ID{}, safeError(ErrAmbiguousID, "raw item/work/lottery id")
	}
	return ID{}, safeError(ErrUnknownLegacyID, "component")
}

func parseRank(raw string) (ID, bool, error) {
	matches := legacyRankRe.FindStringSubmatch(raw)
	pageIndex := 2
	kindIndex := 3
	if matches == nil {
		matches = legacyRankAltRe.FindStringSubmatch(raw)
		pageIndex = 3
		kindIndex = 2
	}
	if matches == nil {
		return ID{}, false, nil
	}
	page, err := strconv.Atoi(matches[pageIndex])
	if err != nil || page < 0 {
		return ID{}, true, safeError(ErrInvalidPayload, "rank page")
	}
	switch matches[kindIndex] {
	case "text_rank":
		return legacyComponent("rank", "text_page", raw), true, nil
	case "voice_rank":
		return legacyComponent("rank", "voice_page", raw), true, nil
	case "coin_rank":
		return legacyComponent("rank", "coin_page", raw), true, nil
	default:
		return ID{}, true, safeError(ErrUnknownLegacyID, "rank")
	}
}

func parseSign(raw string) (ID, bool, error) {
	matches := legacySignRe.FindStringSubmatch(raw)
	if matches == nil {
		return ID{}, false, nil
	}
	month, err := strconv.Atoi(matches[3])
	if err != nil || month < 1 || month > 12 {
		return ID{}, true, safeError(ErrInvalidPayload, "sign month")
	}
	return legacyComponent("economy", "sign_page", raw), true, nil
}

func parseLeaveRole(raw string) (ID, bool, error) {
	matches := legacyLeaveRoleRe.FindStringSubmatch(raw)
	if matches == nil {
		return ID{}, false, nil
	}
	page, err := strconv.Atoi(matches[1])
	if err != nil || page < 0 {
		return ID{}, true, safeError(ErrInvalidPayload, "reward role page")
	}
	if matches[2] == "text_leave_role" {
		return legacyComponent("xp", "text_reward_page", raw), true, nil
	}
	return legacyComponent("xp", "voice_reward_page", raw), true, nil
}

func routeLottery(raw string) bool {
	if !legacyLotteryRe.MatchString(raw) {
		return false
	}
	return strings.HasPrefix(raw, "lotter") || strings.HasSuffix(raw, "search") || strings.HasSuffix(raw, "restart") || strings.HasSuffix(raw, "stop")
}

func legacyComponent(feature, action, raw string) ID {
	return idFromRoute(raw, RouteKey{Kind: InteractionKindComponent, Feature: feature, Action: action, Version: LegacyVersion, Legacy: true})
}

func idFromRoute(raw string, key RouteKey) ID {
	return ID{
		Namespace: Namespace,
		Version:   key.Version,
		Feature:   key.Feature,
		Action:    key.Action,
		Legacy:    key.Legacy,
		Raw:       raw,
		Kind:      key.Kind,
	}
}

func IsSnowflake(value string) bool {
	return legacySnowflakeRe.MatchString(value)
}
