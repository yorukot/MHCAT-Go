package customid

import "regexp"

type ModalField struct {
	CustomID string
	Value    string
}

var (
	legacyVoiceAnswerModalRe  = regexp.MustCompile(`^([0-9]{17,20})anser$`)
	legacyVerificationModalRe = regexp.MustCompile(`^[A-Za-z0-9]{1,16}ver$`)
	legacyWorkCaptchaModalRe  = regexp.MustCompile(`^[0-9]{1,4}$`)
)

func ParseLegacyModal(raw string, fields []ModalField) (ID, error) {
	if raw == "" {
		return ID{}, ErrEmptyID
	}
	if customIDLength(raw) > MaxCustomIDLength {
		return ID{}, safeError(ErrTooLong, "legacy modal")
	}
	if len(fields) == 0 || fields[0].CustomID == "" {
		return ID{}, safeError(ErrUnknownLegacyID, "modal missing first field")
	}
	first := fields[0].CustomID
	switch {
	case raw == "nal" && first == "anntag":
		return legacyModal("announcement", "submit", raw, fields), nil
	case raw == "nal" && first == "ticketcolor":
		return legacyModal("ticket", "panel_submit", raw, fields), nil
	case raw == "nal" && (first == "leave_msgcolor" || first == "leave_msgcontent"):
		return legacyModal("welcome", "leave_submit", raw, fields), nil
	case raw == "nal" && (first == "join_msgcolor" || first == "join_msgcontent"):
		return legacyModal("welcome", "join_legacy", raw, fields), nil
	case raw == "nal" && legacyRoleAddField(first):
		return legacyModal("role_button", "modal_submit", raw, fields), nil
	case first == "cron_setcron":
		return legacyModal("cron", "submit", raw, fields), nil
	case legacyVoiceAnswerModalRe.MatchString(raw) && first == "anser":
		return legacyModal("voice_lock", "answer", raw, fields), nil
	case legacyVerificationModalRe.MatchString(raw) && first == raw:
		return legacyModal("verification", "answer", raw, fields), nil
	case legacyWorkCaptchaModalRe.MatchString(raw) && first == "captcha":
		return legacyModal("work", "captcha", raw, fields), nil
	default:
		return ID{}, safeError(ErrUnknownLegacyID, "modal")
	}
}

func legacyModal(feature, action, raw string, fields []ModalField) ID {
	fieldMap := make(map[string]string, len(fields))
	for _, field := range fields {
		if field.CustomID != "" {
			fieldMap[field.CustomID] = field.Value
		}
	}
	return ID{
		Namespace: Namespace,
		Version:   LegacyVersion,
		Feature:   feature,
		Action:    action,
		Legacy:    true,
		Raw:       raw,
		Kind:      InteractionKindModal,
		Fields:    fieldMap,
	}
}

func legacyRoleAddField(field string) bool {
	return legacyRoleAddModalRe.MatchString(field)
}
