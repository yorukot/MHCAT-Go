package logging

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
)

const loggingChannelUpdateAuditAction = 11

const (
	loggingPermissionCreateInstantInvite int64 = 1 << 0
	loggingPermissionKickMembers         int64 = 1 << 1
	loggingPermissionBanMembers          int64 = 1 << 2
	loggingPermissionAdministrator       int64 = 1 << 3
	loggingPermissionManageChannels      int64 = 1 << 4
	loggingPermissionManageGuild         int64 = 1 << 5
	loggingPermissionAddReactions        int64 = 1 << 6
	loggingPermissionViewAuditLogs       int64 = 1 << 7
	loggingPermissionPrioritySpeaker     int64 = 1 << 8
	loggingPermissionStream              int64 = 1 << 9
	loggingPermissionViewChannel         int64 = 1 << 10
	loggingPermissionSendMessages        int64 = 1 << 11
	loggingPermissionSendTTSMessages     int64 = 1 << 12
	loggingPermissionManageMessages      int64 = 1 << 13
	loggingPermissionEmbedLinks          int64 = 1 << 14
	loggingPermissionAttachFiles         int64 = 1 << 15
	loggingPermissionReadMessageHistory  int64 = 1 << 16
	loggingPermissionMentionEveryone     int64 = 1 << 17
	loggingPermissionUseExternalEmojis   int64 = 1 << 18
	loggingPermissionViewGuildInsights   int64 = 1 << 19
	loggingPermissionConnect             int64 = 1 << 20
	loggingPermissionSpeak               int64 = 1 << 21
	loggingPermissionMuteMembers         int64 = 1 << 22
	loggingPermissionDeafenMembers       int64 = 1 << 23
	loggingPermissionMoveMembers         int64 = 1 << 24
	loggingPermissionUseVAD              int64 = 1 << 25
	loggingPermissionChangeNickname      int64 = 1 << 26
	loggingPermissionManageNicknames     int64 = 1 << 27
	loggingPermissionManageRoles         int64 = 1 << 28
	loggingPermissionManageWebhooks      int64 = 1 << 29
	loggingPermissionManageExpressions   int64 = 1 << 30
	loggingPermissionUseAppCommands      int64 = 1 << 31
	loggingPermissionRequestToSpeak      int64 = 1 << 32
	loggingPermissionManageEvents        int64 = 1 << 33
	loggingPermissionManageThreads       int64 = 1 << 34
	loggingPermissionCreatePublicThreads int64 = 1 << 35
	loggingPermissionCreatePrivateThread int64 = 1 << 36
	loggingPermissionUseExternalStickers int64 = 1 << 37
	loggingPermissionSendInThreads       int64 = 1 << 38
	loggingPermissionUseActivities       int64 = 1 << 39
	loggingPermissionModerateMembers     int64 = 1 << 40
)

type loggingPermissionName struct {
	Name string
	Bit  int64
}

var loggingLegacyPermissionOrder = []loggingPermissionName{
	{Name: "Create Instant Invite", Bit: loggingPermissionCreateInstantInvite},
	{Name: "Kick Members", Bit: loggingPermissionKickMembers},
	{Name: "Ban Members", Bit: loggingPermissionBanMembers},
	{Name: "Administrator", Bit: loggingPermissionAdministrator},
	{Name: "Manage Channels", Bit: loggingPermissionManageChannels},
	{Name: "Manage Guild", Bit: loggingPermissionManageGuild},
	{Name: "Add Reactions", Bit: loggingPermissionAddReactions},
	{Name: "View AuditLog", Bit: loggingPermissionViewAuditLogs},
	{Name: "Priority Speaker", Bit: loggingPermissionPrioritySpeaker},
	{Name: "Stream", Bit: loggingPermissionStream},
	{Name: "View Channel", Bit: loggingPermissionViewChannel},
	{Name: "Send Messages", Bit: loggingPermissionSendMessages},
	{Name: "Send TTS Messages", Bit: loggingPermissionSendTTSMessages},
	{Name: "Manage Messages", Bit: loggingPermissionManageMessages},
	{Name: "Embed Links", Bit: loggingPermissionEmbedLinks},
	{Name: "Attach Files", Bit: loggingPermissionAttachFiles},
	{Name: "Read Message History", Bit: loggingPermissionReadMessageHistory},
	{Name: "Mention Everyone", Bit: loggingPermissionMentionEveryone},
	{Name: "Use External Emojis", Bit: loggingPermissionUseExternalEmojis},
	{Name: "View Guild Insights", Bit: loggingPermissionViewGuildInsights},
	{Name: "Connect", Bit: loggingPermissionConnect},
	{Name: "Speak", Bit: loggingPermissionSpeak},
	{Name: "Mute Members", Bit: loggingPermissionMuteMembers},
	{Name: "Deafen Members", Bit: loggingPermissionDeafenMembers},
	{Name: "Move Members", Bit: loggingPermissionMoveMembers},
	{Name: "Use VAD", Bit: loggingPermissionUseVAD},
	{Name: "Change Nickname", Bit: loggingPermissionChangeNickname},
	{Name: "Manage Nicknames", Bit: loggingPermissionManageNicknames},
	{Name: "Manage Roles", Bit: loggingPermissionManageRoles},
	{Name: "Manage Webhooks", Bit: loggingPermissionManageWebhooks},
	{Name: "Manage Emojis And Stickers", Bit: loggingPermissionManageExpressions},
	{Name: "Use Application Commands", Bit: loggingPermissionUseAppCommands},
	{Name: "Request To Speak", Bit: loggingPermissionRequestToSpeak},
	{Name: "Manage Events", Bit: loggingPermissionManageEvents},
	{Name: "Manage Threads", Bit: loggingPermissionManageThreads},
	{Name: "Create Public Threads", Bit: loggingPermissionCreatePublicThreads},
	{Name: "Create Private Threads", Bit: loggingPermissionCreatePrivateThread},
	{Name: "Use External Stickers", Bit: loggingPermissionUseExternalStickers},
	{Name: "Send Messages In Threads", Bit: loggingPermissionSendInThreads},
	{Name: "Use Embedded Activities", Bit: loggingPermissionUseActivities},
	{Name: "Moderate Members", Bit: loggingPermissionModerateMembers},
}

type loggingPermissionDiff struct {
	Overwrite events.PermissionOverwrite
	Defaults  []string
	Allows    []string
	Denies    []string
}

func NewChannelEventModule(repo ports.LoggingConfigReader, messages ports.DiscordMessagePort, auditLogs ports.DiscordAuditLogPort) Module {
	return Module{
		configReader:         repo,
		messages:             messages,
		auditLogs:            auditLogs,
		channelEventsEnabled: repo != nil && messages != nil,
	}
}

func (m Module) ChannelUpdateHandler() events.Handler {
	return func(ctx context.Context, event events.Event) error {
		if event.Type != events.TypeChannelUpdate || event.ChannelUpdate == nil || !event.ChannelUpdate.HasOldChannel {
			return nil
		}
		channel := event.ChannelUpdate
		if strings.TrimSpace(event.GuildID) == "" || strings.TrimSpace(event.ChannelID) == "" {
			return nil
		}
		config, err := m.configReader.GetLoggingConfig(ctx, event.GuildID)
		if err != nil {
			if errors.Is(err, ports.ErrLoggingConfigMissing) {
				return nil
			}
			return err
		}
		if !config.ChannelUpdate || strings.TrimSpace(config.ChannelID) == "" {
			return nil
		}
		if channel.OldTopic != channel.NewTopic || channel.OldTopicNull != channel.NewTopicNull {
			actor := m.channelUpdateActor(ctx, event)
			_, err = m.messages.SendMessage(ctx, config.ChannelID, ports.OutboundMessage{
				Embeds: []ports.OutboundEmbed{{
					AuthorName:    loggingAuditUsername(actor) + " | 頻道主題更新",
					AuthorIconURL: loggingAuditAvatarURL(actor),
					Color:         0xFF8040,
					Description:   "**<:chat:1085254765342109697> 頻道主題編輯者: <@" + actor.UserID + "> | <:Channel:994524759289233438> 頻道: <#" + event.ChannelID + ">**",
					Fields: []ports.OutboundEmbedField{
						{Name: "**<:book:1084846007545778217> 舊主題**", Value: loggingCodeBlock(loggingTopicText(channel.OldTopic, channel.OldTopicNull))},
						{Name: "**<:new:1084846011366785135> 新主題:**", Value: loggingCodeBlock(loggingTopicText(channel.NewTopic, channel.NewTopicNull))},
					},
					FooterText:    loggingFooterText,
					FooterIconURL: event.BotAvatarURL,
					Timestamp:     time.Now(),
				}},
				AllowedMentions: ports.AllowedMentions{},
			})
			return err
		}
		diffs := loggingPermissionDiffs(*channel)
		if len(diffs) == 0 {
			return nil
		}
		actor := m.channelUpdateActor(ctx, event)
		for _, diff := range diffs {
			_, err := m.messages.SendMessage(ctx, config.ChannelID, ports.OutboundMessage{
				Embeds: []ports.OutboundEmbed{{
					AuthorName:    loggingAuditUsername(actor) + " | 頻道權限更新",
					AuthorIconURL: loggingAuditAvatarURL(actor),
					Color:         0xFF5809,
					Description:   "**<:shield:1019529265101930567> 頻道權限編輯者: <@" + actor.UserID + "> | <:Channel:994524759289233438> 頻道: <#" + event.ChannelID + ">**",
					Fields: []ports.OutboundEmbedField{{
						Name:  "**<:roleplaying:985945121264635964> 身分組或使用者: **",
						Value: loggingPermissionFieldValue(diff),
					}},
					FooterText:    loggingFooterText,
					FooterIconURL: event.BotAvatarURL,
					Timestamp:     time.Now(),
				}},
				AllowedMentions: ports.AllowedMentions{},
			})
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func loggingTopicText(topic string, isNull bool) string {
	if isNull {
		return "null"
	}
	return topic
}

func (m Module) channelUpdateActor(ctx context.Context, event events.Event) ports.AuditLogEntry {
	if m.auditLogs == nil {
		return ports.AuditLogEntry{}
	}
	entries, err := m.auditLogs.AuditLog(ctx, ports.AuditLogQuery{
		GuildID: event.GuildID,
		Action:  loggingChannelUpdateAuditAction,
		Limit:   1,
	})
	if err != nil {
		return ports.AuditLogEntry{}
	}
	if len(entries) > 0 {
		return entries[0]
	}
	return ports.AuditLogEntry{}
}

func loggingPermissionDiffs(update events.ChannelUpdate) []loggingPermissionDiff {
	oldByID := map[string]events.PermissionOverwrite{}
	for _, overwrite := range update.OldPermissionOverwrites {
		if strings.TrimSpace(overwrite.ID) != "" {
			oldByID[overwrite.ID] = overwrite
		}
	}
	diffs := []loggingPermissionDiff{}
	for _, current := range update.NewPermissionOverwrites {
		if strings.TrimSpace(current.ID) == "" {
			continue
		}
		old, hadOld := oldByID[current.ID]
		diff := loggingPermissionDiff{Overwrite: current}
		for _, permission := range loggingLegacyPermissionOrder {
			currentAllow := current.Allow&permission.Bit != 0
			oldAllow := hadOld && old.Allow&permission.Bit != 0
			if currentAllow != oldAllow {
				if !currentAllow && oldAllow {
					diff.Defaults = append(diff.Defaults, "<:YellowSmallDot:1023970607429328946> "+permission.Name)
				} else {
					diff.Allows = append(diff.Allows, "<:check:1085240252978966548> "+permission.Name)
				}
			}
			currentDeny := current.Deny&permission.Bit != 0
			oldDeny := hadOld && old.Deny&permission.Bit != 0
			if currentDeny != oldDeny {
				if !currentDeny && oldDeny {
					diff.Defaults = append(diff.Defaults, "<:YellowSmallDot:1023970607429328946> "+permission.Name)
				} else {
					diff.Denies = append(diff.Denies, "<:prohibition:1085240255810129960> "+permission.Name)
				}
			}
		}
		if len(diff.Defaults) != 0 || len(diff.Allows) != 0 || len(diff.Denies) != 0 {
			diffs = append(diffs, diff)
		}
	}
	return diffs
}

func loggingPermissionFieldValue(diff loggingPermissionDiff) string {
	return "<:icons_text1:1000814305068986590>" + loggingOverwriteMention(diff.Overwrite) + "\n" +
		strings.Join(diff.Defaults, "\n") + "\n" +
		strings.Join(diff.Allows, "\n") + "\n" +
		strings.Join(diff.Denies, "\n")
}

func loggingOverwriteMention(overwrite events.PermissionOverwrite) string {
	if overwrite.Type == 0 {
		return "<@&" + overwrite.ID + ">"
	}
	return "<@" + overwrite.ID + ">"
}

func loggingAuditUsername(entry ports.AuditLogEntry) string {
	if strings.TrimSpace(entry.Username) != "" {
		return strings.TrimSpace(entry.Username)
	}
	if strings.TrimSpace(entry.UserTag) != "" {
		if before, _, ok := strings.Cut(strings.TrimSpace(entry.UserTag), "#"); ok && before != "" {
			return before
		}
		return strings.TrimSpace(entry.UserTag)
	}
	if strings.TrimSpace(entry.UserID) != "" {
		return entry.UserID
	}
	return "unknown"
}

func loggingAuditAvatarURL(entry ports.AuditLogEntry) string {
	raw := strings.TrimSpace(entry.AvatarURL)
	if raw != "" {
		return loggingPNGAvatarURL(raw)
	}
	return loggingDefaultAvatarURL
}
