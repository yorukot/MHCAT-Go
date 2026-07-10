package voice

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/voice"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	permissionManageMessages  = int64(8192)
	voiceErrorColor           = 0xED4245
	voiceSuccessColor         = 0x57F287
	legacyVoiceLockColor      = 0x53FF53
	legacyVoiceLockErrorColor = 0xEA0000
	legacyVoiceEmoji          = "<:Voice:994844272790610011>"
	legacyDoneEmoji           = "<a:green_tick:994529015652163614>"
	legacyDeleteEmoji         = "<:trashbin:995991389043163257>"
	discordChannelTypeVoice   = 2
	permissionOverwriteMember = 1
	permissionManageChannels  = 16
	permissionManageRoles     = 268435456
	voiceLockAnswerInputID    = "anser"
	legacyUnlockEmoji         = "<:unlock:1017087850556174367>"
	legacyLockEmoji           = "<:lock:1017077025397288980>"
	legacyArrowPinkEmoji      = "<a:arrow_pink:996242460294512690>"
	voiceLockPromptTTL        = 60 * time.Second
)

func (m Module) SetHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, voiceErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		trigger := firstCommandOption(interaction, optionTriggerChannel)
		triggerID := strings.TrimSpace(trigger.String)
		limit, ok := limitOption(interaction, optionUserLimit)
		if ok && (limit < 0 || limit > 99) {
			return responder.EditOriginal(ctx, voiceErrorMessage("必須為1-99的整數!"))
		}
		config := domain.VoiceRoomConfig{
			GuildID:          interaction.Actor.GuildID,
			TriggerChannelID: triggerID,
			ParentID:         trigger.ChannelParentID,
			Name:             rawOption(interaction, optionRoomName),
			Limit:            limit,
			Lock:             boolOption(interaction, optionOwnerLock),
		}
		if err := m.service.Save(ctx, config); err != nil {
			return responder.EditOriginal(ctx, voiceUnknownError(err))
		}
		if err := responder.EditOriginal(ctx, voiceSetSuccessMessage()); err != nil {
			return err
		}
		return m.track(ctx, interaction, VoiceRoomSetCommandName)
	}
}

func (m Module) DeleteHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, voiceErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		selected := firstCommandOption(interaction, optionChannelOrGroup)
		channelID := strings.TrimSpace(selected.String)
		var err error
		if selected.ChannelType == discordChannelTypeVoice {
			err = m.service.DeleteByTrigger(ctx, interaction.Actor.GuildID, channelID)
			if err != nil {
				if errors.Is(err, ports.ErrVoiceRoomConfigMissing) {
					return responder.EditOriginal(ctx, voiceErrorMessage("你沒有對這個頻道做出設定過喔!"))
				}
				return responder.EditOriginal(ctx, voiceUnknownError(err))
			}
			if err := responder.EditOriginal(ctx, voiceDeleteTriggerSuccessMessage()); err != nil {
				return err
			}
		} else {
			err = m.service.DeleteByParent(ctx, interaction.Actor.GuildID, channelID)
			if err != nil {
				if errors.Is(err, ports.ErrVoiceRoomConfigMissing) {
					return responder.EditOriginal(ctx, voiceErrorMessage("你沒有對這個類別沒有設定喔!"))
				}
				return responder.EditOriginal(ctx, voiceUnknownError(err))
			}
			if err := responder.EditOriginal(ctx, voiceDeleteParentSuccessMessage()); err != nil {
				return err
			}
		}
		return m.track(ctx, interaction, VoiceRoomDeleteCommandName)
	}
}

func (m LockModule) LockHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		voiceChannelID := strings.TrimSpace(interaction.Actor.VoiceChannelID)
		if voiceChannelID == "" {
			return responder.EditOriginal(ctx, voiceLockErrorMessage("你不在一個語音包廂!"))
		}
		password := rawOption(interaction, optionLockPassword)
		err := m.service.SetPassword(ctx, interaction.Actor.GuildID, voiceChannelID, interaction.Actor.UserID, interaction.ChannelID, password)
		if err != nil {
			switch {
			case errors.Is(err, ports.ErrVoiceRoomLockMissing):
				return responder.EditOriginal(ctx, voiceLockErrorMessage("你不在語音包廂或該語音包廂不支援設密碼!"))
			case errors.Is(err, ports.ErrVoiceRoomLockNotOwner):
				return responder.EditOriginal(ctx, voiceLockErrorMessage("只有包廂房主可以設定!"))
			default:
				return responder.EditOriginal(ctx, voiceUnknownError(err))
			}
		}
		if err := responder.EditOriginal(ctx, voiceLockSuccessMessage(password)); err != nil {
			return err
		}
		return m.track(ctx, interaction, VoiceRoomLockCommandName)
	}
}

func (m LockModule) AnswerHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		channelID := voiceLockAnswerChannelID(interaction.CustomID)
		password := voiceLockAnswerValue(interaction.ModalFields)
		err := m.service.AnswerPassword(ctx, interaction.Actor.GuildID, channelID, interaction.Actor.UserID, password)
		if err != nil {
			switch {
			case errors.Is(err, coreservice.ErrVoiceRoomLockPasswordMismatch):
				return responder.EditOriginal(ctx, voiceLockAnswerWrongMessage(interaction.Actor.GuildID, channelID))
			case errors.Is(err, ports.ErrVoiceRoomLockMissing):
				return responder.EditOriginal(ctx, voiceLockAnswerMissingMessage())
			default:
				return responder.EditOriginal(ctx, voiceUnknownError(err))
			}
		}
		return responder.EditOriginal(ctx, voiceLockAnswerSuccessMessage(interaction.Actor.GuildID, channelID))
	}
}

func (m LockModule) PromptHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		channelID, userID, expiresAt, ok := voiceLockPromptPayload(interaction.CustomID)
		if !ok || !m.now().Before(expiresAt) {
			message := voiceLockPromptUnavailableMessage()
			message.Ephemeral = true
			return responder.Reply(ctx, message)
		}
		if userID != "" && strings.TrimSpace(interaction.Actor.UserID) != userID {
			message := voiceLockPromptUnavailableMessage()
			message.Ephemeral = true
			return responder.Reply(ctx, message)
		}
		return responder.ShowModal(ctx, voiceLockAnswerModal(channelID))
	}
}

func (m LockEventModule) VoiceStateHandler() events.Handler {
	return func(ctx context.Context, event events.Event) error {
		if event.Type != events.TypeVoiceState || event.VoiceState == nil {
			return nil
		}
		voice := event.VoiceState
		guildID := strings.TrimSpace(voice.GuildID)
		if guildID == "" {
			guildID = strings.TrimSpace(event.GuildID)
		}
		channelID := strings.TrimSpace(voice.ChannelID)
		userID := strings.TrimSpace(voice.UserID)
		if userID == "" {
			userID = strings.TrimSpace(event.UserID)
		}
		isBot := event.IsBot
		if event.Member != nil {
			isBot = event.Member.IsBot
			if event.Member.UserID != "" {
				userID = strings.TrimSpace(event.Member.UserID)
			}
		}
		if guildID == "" || channelID == "" || userID == "" || isBot || channelID == strings.TrimSpace(voice.BeforeChannel) {
			return nil
		}
		lock, prompt, err := m.service.LockedJoinPrompt(ctx, guildID, channelID, userID)
		if err != nil || !prompt {
			return err
		}
		expiresAt := m.now().Add(voiceLockPromptTTL)
		if _, err := m.messages.SendMessage(ctx, lock.TextChannelID, voiceLockPromptOutbound(lock, userID, expiresAt)); err != nil {
			return err
		}
		if err := m.members.MoveMember(ctx, guildID, userID, nil); err != nil {
			return err
		}
		_, _ = m.direct.SendDirectMessage(ctx, userID, voiceLockDirectMessage(lock, m.randomColor()))
		return ctx.Err()
	}
}

func (m RoomEventModule) VoiceStateHandler() events.Handler {
	return func(ctx context.Context, event events.Event) error {
		if event.Type != events.TypeVoiceState || event.VoiceState == nil {
			return nil
		}
		voice := event.VoiceState
		guildID := strings.TrimSpace(voice.GuildID)
		if guildID == "" {
			guildID = strings.TrimSpace(event.GuildID)
		}
		channelID := strings.TrimSpace(voice.ChannelID)
		beforeChannelID := strings.TrimSpace(voice.BeforeChannel)
		userID := strings.TrimSpace(voice.UserID)
		if userID == "" {
			userID = strings.TrimSpace(event.UserID)
		}
		username := event.UserTag
		if event.Member != nil {
			if event.Member.UserID != "" {
				userID = strings.TrimSpace(event.Member.UserID)
			}
			if event.Member.Username != "" {
				username = event.Member.Username
			} else if event.Member.UserTag != "" {
				username = event.Member.UserTag
			}
		}
		if guildID == "" || channelID == beforeChannelID {
			return nil
		}
		if channelID != "" && userID != "" {
			if err := m.createDynamicRoom(ctx, guildID, channelID, userID, username); err != nil {
				return err
			}
		}
		if beforeChannelID == "" {
			return ctx.Err()
		}
		return m.cleanupDynamicRoom(ctx, guildID, beforeChannelID)
	}
}

func (m RoomEventModule) createDynamicRoom(ctx context.Context, guildID string, triggerChannelID string, userID string, username string) error {
	config, ok, err := m.service.TriggerConfig(ctx, guildID, triggerChannelID)
	if err != nil || !ok {
		return err
	}
	overwrites, err := m.dynamicRoomPermissionOverwrites(ctx, guildID, config.ParentID, userID)
	if err != nil {
		return err
	}
	created, err := m.channels.CreateChannel(ctx, ports.ChannelCreateRequest{
		GuildID:              guildID,
		ParentID:             config.ParentID,
		Name:                 voiceRoomDynamicName(config.Name, username, userID),
		Type:                 discordChannelTypeVoice,
		UserLimit:            config.Limit,
		PermissionOverwrites: overwrites,
	})
	if err != nil {
		return err
	}
	if err := m.service.TrackDynamicRoom(ctx, guildID, created.ChannelID, userID, config.Lock); err != nil {
		_ = m.channels.DeleteChannel(context.Background(), created.ChannelID)
		return err
	}
	if err := m.members.MoveMember(ctx, guildID, userID, &created.ChannelID); err != nil {
		_ = m.service.DeleteDynamicRoomLock(context.Background(), guildID, created.ChannelID)
		_ = m.channels.DeleteChannel(context.Background(), created.ChannelID)
		_ = m.service.DeleteDynamicRoomState(context.Background(), guildID, created.ChannelID)
		return err
	}
	if config.Lock {
		_, _ = m.direct.SendDirectMessage(ctx, userID, voiceRoomLockableOwnerMessage(m.randomColor()))
	}
	return ctx.Err()
}

func (m RoomEventModule) cleanupDynamicRoom(ctx context.Context, guildID string, channelID string) error {
	tracked, err := m.service.IsDynamicRoom(ctx, guildID, channelID)
	if err != nil || !tracked {
		return err
	}
	memberCount, err := m.channels.VoiceChannelMemberCount(ctx, guildID, channelID)
	if err != nil && !errors.Is(err, ports.ErrChannelNotFound) {
		return err
	}
	if memberCount > 0 {
		return ctx.Err()
	}
	if err := m.service.DeleteDynamicRoomLock(ctx, guildID, channelID); err != nil {
		return err
	}
	if err := m.channels.DeleteChannel(ctx, channelID); err != nil && !errors.Is(err, ports.ErrChannelNotFound) {
		return err
	}
	return m.service.DeleteDynamicRoomState(ctx, guildID, channelID)
}

func (m RoomEventModule) dynamicRoomPermissionOverwrites(ctx context.Context, guildID string, parentID string, userID string) ([]ports.PermissionOverwrite, error) {
	overwrites := []ports.PermissionOverwrite{}
	if strings.TrimSpace(parentID) != "" {
		parent, err := m.channels.FindChannelByID(ctx, guildID, parentID)
		if err != nil && !errors.Is(err, ports.ErrChannelNotFound) {
			return nil, err
		}
		overwrites = append(overwrites, parent.PermissionOverwrites...)
	}
	return upsertOwnerOverwrite(overwrites, userID), nil
}

func voiceSetSuccessMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       legacyDoneEmoji + " | 成功進行設定",
			Description: legacyVoiceEmoji + " 你成功對語音包廂進行`設定`",
			Color:       voiceSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func voiceDeleteTriggerSuccessMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       legacyDoneEmoji + "成功進行刪除",
			Description: legacyDeleteEmoji + "你成功對這個設定刪除",
			Color:       voiceSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func voiceDeleteParentSuccessMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "成功進行刪除",
			Description: "你成功對這個設定刪除",
			Color:       voiceSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func voiceUnknownError(err error) responses.Message {
	_ = err
	return voiceErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
}

func voiceErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: voiceErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func voiceLockErrorMessage(content string) responses.Message {
	message := voiceErrorMessage(content)
	message.Ephemeral = true
	return message
}

func voiceLockSuccessMessage(password string) responses.Message {
	if password == "" {
		password = "null"
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       legacyDoneEmoji + " | 成功進行設定",
			Description: legacyVoiceEmoji + " 你成功對語音包廂密碼進行設定為:" + password,
			Color:       voiceSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
		Ephemeral:       true,
	}
}

func voiceLockAnswerModal(channelID string) responses.Modal {
	return responses.Modal{
		CustomID: strings.TrimSpace(channelID) + "anser",
		Title:    "請輸入密碼!",
		Rows: []responses.ModalRow{{
			Inputs: []responses.TextInput{{
				CustomID: voiceLockAnswerInputID,
				Label:    "請輸入包廂密碼!",
				Style:    responses.TextInputStyleShort,
				Required: true,
			}},
		}},
	}
}

func voiceLockAnswerSuccessMessage(guildID string, channelID string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: legacyUnlockEmoji + " | 您成功輸入正確密碼\n可以重新加入語音頻道囉!",
			Color: legacyVoiceLockColor,
		}},
		Components:      []responses.ComponentRow{voiceLockChannelLinkRow(guildID, channelID)},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func voiceLockAnswerWrongMessage(guildID string, channelID string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | 你的密碼輸入錯誤!請重新加入語音頻道後在試一次!",
			Color: legacyVoiceLockErrorColor,
		}},
		Components:      []responses.ComponentRow{voiceLockChannelLinkRow(guildID, channelID)},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func voiceLockAnswerMissingMessage() responses.Message {
	return voiceErrorMessage("很抱歉，該包廂可能已被刪除!")
}

func voiceLockPromptUnavailableMessage() responses.Message {
	return voiceErrorMessage("請重新加入語音頻道後再試一次!")
}

func voiceLockChannelLinkRow(guildID string, channelID string) responses.ComponentRow {
	return responses.ComponentRow{Components: []responses.Component{{
		Type:  responses.ComponentTypeButton,
		Label: "點我前往該語音頻道!",
		Emoji: legacyArrowPinkEmoji,
		URL:   "https://discord.com/channels/" + strings.TrimSpace(guildID) + "/" + strings.TrimSpace(channelID),
		Style: responses.ButtonStyleLink,
	}}}
}

func voiceLockPromptOutbound(lock domain.VoiceRoomLock, userID string, expiresAt time.Time) ports.OutboundMessage {
	return ports.OutboundMessage{
		Content: "<@" + strings.TrimSpace(userID) + ">",
		Embeds: []ports.OutboundEmbed{{
			Title:       legacyUnlockEmoji + " | 該包廂已被上鎖，請輸入密碼",
			Description: "請於60秒內點選下面的按鈕\n輸入密碼來加入<#" + strings.TrimSpace(lock.ChannelID) + ">(請找房主拿密碼)",
			Color:       legacyVoiceLockColor,
		}},
		Components: []ports.OutboundComponentRow{{
			Components: []ports.OutboundComponent{{
				Type:     "button",
				CustomID: voiceLockPromptButtonID(lock.ChannelID, userID, expiresAt),
				Label:    "點我輸入密碼!",
				Emoji:    legacyArrowPinkEmoji,
				Style:    "success",
			}},
		}},
		AllowedMentions: ports.AllowedMentions{UserIDs: []string{strings.TrimSpace(userID)}},
	}
}

func voiceLockDirectMessage(lock domain.VoiceRoomLock, color int) ports.OutboundMessage {
	return ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title:       legacyLockEmoji + " | 該語音頻道已被房主上鎖!",
			Description: "請前往<#" + strings.TrimSpace(lock.TextChannelID) + ">輸入密碼進行解鎖\n否則你將無法加入\n輸入完密碼後就可以重新加入囉!",
			Color:       color,
		}},
		AllowedMentions: ports.AllowedMentions{},
	}
}

func voiceRoomLockableOwnerMessage(color int) ports.OutboundMessage {
	return ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title:       legacyLockEmoji + " | 你開啟了一個可上鎖的語音頻道!",
			Description: "**你可以到你所在的語音頻道伺服器\n在該伺服器打指令的頻道打上**`/上鎖頻道 密碼:`\n**當然也可以不用上鎖\n如需解除上鎖只需打**`/上鎖頻道`\n**當頻道上鎖後對方將會被踢\n並且傳送密碼輸入給該名使用者\n對方輸入正確密碼後即可解鎖**",
			Color:       color,
		}},
		AllowedMentions: ports.AllowedMentions{},
	}
}

func voiceRoomDynamicName(template string, username string, userID string) string {
	name := template
	replacement := username
	if replacement == "" {
		replacement = strings.TrimSpace(userID)
	}
	if replacement == "" {
		replacement = "user"
	}
	if name == "" {
		return replacement
	}
	return strings.Replace(name, "{name}", replacement, 1)
}

func upsertOwnerOverwrite(overwrites []ports.PermissionOverwrite, userID string) []ports.PermissionOverwrite {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return overwrites
	}
	allow := int64(permissionManageChannels | permissionManageRoles)
	out := append([]ports.PermissionOverwrite(nil), overwrites...)
	for index := range out {
		if out[index].ID == userID && out[index].Type == permissionOverwriteMember {
			out[index].Allow |= allow
			out[index].Deny &^= allow
			return out
		}
	}
	return append(out, ports.PermissionOverwrite{ID: userID, Type: permissionOverwriteMember, Allow: allow})
}

func firstCommandOption(interaction interactions.Interaction, name string) interactions.CommandOptionValue {
	if value, ok := interaction.CommandOptions[name]; ok {
		if strings.TrimSpace(value.String) != "" {
			return value
		}
	}
	return interactions.CommandOptionValue{
		Type:   interactions.CommandOptionChannel,
		String: strings.TrimSpace(interaction.Options[name]),
	}
}

func firstOption(interaction interactions.Interaction, name string) string {
	if value := strings.TrimSpace(interaction.Options[name]); value != "" {
		return value
	}
	if option, ok := interaction.CommandOptions[name]; ok {
		return strings.TrimSpace(option.String)
	}
	return ""
}

func rawOption(interaction interactions.Interaction, name string) string {
	if option, ok := interaction.CommandOptions[name]; ok {
		return option.String
	}
	return interaction.Options[name]
}

func voiceLockAnswerChannelID(customID string) string {
	return strings.TrimSpace(strings.TrimSuffix(customID, "anser"))
}

func voiceLockPromptButtonID(channelID string, userID string, expiresAt time.Time) string {
	payload, err := customid.KeyValuePayload(map[string]string{
		"c": strings.TrimSpace(channelID),
		"e": strconv.FormatInt(expiresAt.UnixMilli(), 10),
		"u": strings.TrimSpace(userID),
	})
	if err != nil {
		return "lock_start"
	}
	id, err := customid.Encode(customid.InteractionKindComponent, "voice_lock", "prompt", payload)
	if err != nil {
		return "lock_start"
	}
	return id
}

func voiceLockPromptPayload(raw string) (string, string, time.Time, bool) {
	parsed, err := customid.ParseComponent(raw)
	if err != nil || parsed.Feature != "voice_lock" || parsed.Action != "prompt" || parsed.Payload.Kind != customid.PayloadKV {
		return "", "", time.Time{}, false
	}
	channelID := strings.TrimSpace(parsed.Payload.Values["c"])
	userID := strings.TrimSpace(parsed.Payload.Values["u"])
	expiresMillis, err := strconv.ParseInt(parsed.Payload.Values["e"], 10, 64)
	if channelID == "" || userID == "" || err != nil || expiresMillis <= 0 {
		return "", "", time.Time{}, false
	}
	return channelID, userID, time.UnixMilli(expiresMillis), true
}

func (m LockModule) now() time.Time {
	if m.clock == nil {
		return time.Now()
	}
	return m.clock.Now()
}

func (m LockEventModule) now() time.Time {
	if m.clock == nil {
		return time.Now()
	}
	return m.clock.Now()
}

func (m LockEventModule) randomColor() int {
	if m.color == nil {
		return legacyVoiceRandomColor()
	}
	return m.color()
}

func (m RoomEventModule) randomColor() int {
	if m.color == nil {
		return legacyVoiceRandomColor()
	}
	return m.color()
}

func voiceLockAnswerValue(fields []customid.ModalField) string {
	for _, field := range fields {
		if field.CustomID == voiceLockAnswerInputID {
			return field.Value
		}
	}
	if len(fields) == 0 {
		return ""
	}
	return fields[0].Value
}

func boolOption(interaction interactions.Interaction, name string) bool {
	if option, ok := interaction.CommandOptions[name]; ok {
		return option.Bool
	}
	return strings.EqualFold(strings.TrimSpace(interaction.Options[name]), "true")
}

func limitOption(interaction interactions.Interaction, name string) (int, bool) {
	if option, ok := interaction.CommandOptions[name]; ok {
		return int(option.Int), true
	}
	raw := strings.TrimSpace(interaction.Options[name])
	if raw == "" {
		return 0, false
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return 0, true
	}
	return parsed, true
}

func (m Module) track(ctx context.Context, interaction interactions.Interaction, commandName string) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: commandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     "voice-room-config",
	})
}

func (m LockModule) track(ctx context.Context, interaction interactions.Interaction, commandName string) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: commandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     "voice-room-lock",
	})
}
