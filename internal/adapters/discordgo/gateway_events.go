package discordgo

import (
	"context"
	"time"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
)

type GatewayEventOptions struct {
	Messages         bool
	GuildChannels    bool
	MessageReactions bool
	GuildMembers     bool
	VoiceStates      bool
}

func discordIDCreatedAt(id string) time.Time {
	createdAt, err := dgo.SnowflakeTimestamp(id)
	if err != nil {
		return time.Time{}
	}
	return zeroIfInvalid(createdAt)
}

func (s *Session) RegisterGatewayEventHandlers(dispatcher *events.Dispatcher, opts GatewayEventOptions) func() {
	if s == nil || s.session == nil || dispatcher == nil {
		return func() {}
	}
	var removers []func()
	removers = append(removers,
		s.session.AddHandler(func(_ *dgo.Session, event *dgo.Ready) {
			dispatcher.DispatchSafe(context.Background(), events.Event{Type: events.TypeReady})
		}),
		s.session.AddHandler(func(_ *dgo.Session, event *dgo.Resumed) {
			dispatcher.DispatchSafe(context.Background(), events.Event{Type: events.TypeResumed})
		}),
	)
	if opts.Messages {
		removers = append(removers,
			s.session.AddHandler(func(session *dgo.Session, event *dgo.MessageCreate) {
				dispatcher.DispatchSafe(context.Background(), eventFromMessage(events.TypeMessageCreate, event.Message, botFromState(session)))
			}),
			s.session.AddHandler(func(session *dgo.Session, event *dgo.MessageUpdate) {
				dispatcher.DispatchSafe(context.Background(), eventFromMessageUpdate(event, botFromState(session)))
			}),
			s.session.AddHandler(func(session *dgo.Session, event *dgo.MessageDelete) {
				dispatcher.DispatchSafe(context.Background(), eventFromMessageDelete(event, botFromState(session)))
			}),
		)
	}
	if opts.GuildChannels {
		removers = append(removers,
			s.session.AddHandler(func(session *dgo.Session, event *dgo.ChannelUpdate) {
				dispatcher.DispatchSafe(context.Background(), eventFromChannelUpdate(event, botFromState(session)))
			}),
		)
	}
	if opts.MessageReactions {
		removers = append(removers,
			s.session.AddHandler(func(_ *dgo.Session, event *dgo.MessageReactionAdd) {
				dispatcher.DispatchSafe(context.Background(), eventFromReaction(events.TypeReactionAdd, event.MessageReaction, event.Member))
			}),
			s.session.AddHandler(func(_ *dgo.Session, event *dgo.MessageReactionRemove) {
				dispatcher.DispatchSafe(context.Background(), eventFromReaction(events.TypeReactionRemove, event.MessageReaction, nil))
			}),
		)
	}
	if opts.GuildMembers {
		removers = append(removers,
			s.session.AddHandler(func(session *dgo.Session, event *dgo.GuildMemberAdd) {
				var member *dgo.Member
				if event != nil {
					member = event.Member
				}
				dispatcher.DispatchSafe(context.Background(), eventFromMember(events.TypeMemberAdd, member, guildFromState(session, member), botFromState(session)))
			}),
			s.session.AddHandler(func(session *dgo.Session, event *dgo.GuildMemberRemove) {
				var member *dgo.Member
				if event != nil {
					member = event.Member
				}
				dispatcher.DispatchSafe(context.Background(), eventFromMember(events.TypeMemberRemove, member, guildFromState(session, member), botFromState(session)))
			}),
		)
	}
	if opts.VoiceStates {
		removers = append(removers, s.session.AddHandler(func(session *dgo.Session, event *dgo.VoiceStateUpdate) {
			dispatcher.DispatchSafe(context.Background(), eventFromVoiceState(event.VoiceState, event.BeforeUpdate, session))
		}))
	}
	return func() {
		for i := len(removers) - 1; i >= 0; i-- {
			if removers[i] != nil {
				removers[i]()
			}
		}
	}
}

func eventFromMessage(eventType events.Type, message *dgo.Message, bot *dgo.User) events.Event {
	if message == nil {
		return events.Event{Type: eventType}
	}
	event := events.Event{
		Type:      eventType,
		ID:        message.ID,
		MessageID: message.ID,
		GuildID:   message.GuildID,
		ChannelID: message.ChannelID,
		Content:   message.Content,
		CreatedAt: message.Timestamp,
	}
	if bot != nil {
		event.BotUserID = bot.ID
		event.BotAvatarURL = bot.AvatarURL("")
	}
	if message.Author != nil {
		event.UserID = message.Author.ID
		event.Username = message.Author.Username
		event.UserTag = userTag(message.Author)
		event.AvatarURL = message.Author.AvatarURL("")
		event.IsBot = message.Author.Bot
	}
	event.Attachments = attachmentsFromDiscord(message.Attachments)
	if message.Member != nil {
		member := *message.Member
		if member.User == nil {
			member.User = message.Author
		}
		event.Member = memberFromDiscord(&member)
	}
	return event
}

func eventFromMessageUpdate(update *dgo.MessageUpdate, bot *dgo.User) events.Event {
	if update == nil {
		return events.Event{Type: events.TypeMessageUpdate}
	}
	event := eventFromMessage(events.TypeMessageUpdate, update.Message, bot)
	if update.BeforeUpdate != nil {
		event.OldContent = update.BeforeUpdate.Content
		event.HasOldContent = true
		if event.UserID == "" && update.BeforeUpdate.Author != nil {
			event.UserID = update.BeforeUpdate.Author.ID
			event.Username = update.BeforeUpdate.Author.Username
			event.UserTag = userTag(update.BeforeUpdate.Author)
			event.AvatarURL = update.BeforeUpdate.Author.AvatarURL("")
			event.IsBot = update.BeforeUpdate.Author.Bot
		}
	}
	return event
}

func eventFromMessageDelete(deleted *dgo.MessageDelete, bot *dgo.User) events.Event {
	if deleted == nil {
		return events.Event{Type: events.TypeMessageDelete}
	}
	message := deleted.Message
	if deleted.BeforeDelete != nil {
		message = deleted.BeforeDelete
		if message.ID == "" && deleted.Message != nil {
			message.ID = deleted.Message.ID
		}
		if message.ChannelID == "" && deleted.Message != nil {
			message.ChannelID = deleted.Message.ChannelID
		}
		if message.GuildID == "" && deleted.Message != nil {
			message.GuildID = deleted.Message.GuildID
		}
	}
	return eventFromMessage(events.TypeMessageDelete, message, bot)
}

func attachmentsFromDiscord(attachments []*dgo.MessageAttachment) []events.Attachment {
	if len(attachments) == 0 {
		return nil
	}
	out := make([]events.Attachment, 0, len(attachments))
	for _, attachment := range attachments {
		if attachment == nil || attachment.URL == "" {
			continue
		}
		out = append(out, events.Attachment{URL: attachment.URL})
	}
	return out
}

func eventFromChannelUpdate(update *dgo.ChannelUpdate, bot *dgo.User) events.Event {
	event := events.Event{Type: events.TypeChannelUpdate}
	if bot != nil {
		event.BotUserID = bot.ID
		event.BotAvatarURL = bot.AvatarURL("")
	}
	channelUpdate := &events.ChannelUpdate{}
	if update != nil && update.Channel != nil {
		event.ID = update.Channel.ID
		event.ChannelID = update.Channel.ID
		event.GuildID = update.Channel.GuildID
		channelUpdate.ChannelID = update.Channel.ID
		channelUpdate.GuildID = update.Channel.GuildID
		channelUpdate.NewTopic = update.Channel.Topic
		channelUpdate.NewPermissionOverwrites = permissionOverwritesFromDiscord(update.Channel.PermissionOverwrites)
	}
	if update != nil && update.BeforeUpdate != nil {
		channelUpdate.HasOldChannel = true
		channelUpdate.OldTopic = update.BeforeUpdate.Topic
		channelUpdate.OldPermissionOverwrites = permissionOverwritesFromDiscord(update.BeforeUpdate.PermissionOverwrites)
		if event.ID == "" {
			event.ID = update.BeforeUpdate.ID
		}
		if event.ChannelID == "" {
			event.ChannelID = update.BeforeUpdate.ID
			channelUpdate.ChannelID = update.BeforeUpdate.ID
		}
		if event.GuildID == "" {
			event.GuildID = update.BeforeUpdate.GuildID
			channelUpdate.GuildID = update.BeforeUpdate.GuildID
		}
	}
	event.ChannelUpdate = channelUpdate
	return event
}

func permissionOverwritesFromDiscord(overwrites []*dgo.PermissionOverwrite) []events.PermissionOverwrite {
	if len(overwrites) == 0 {
		return nil
	}
	out := make([]events.PermissionOverwrite, 0, len(overwrites))
	for _, overwrite := range overwrites {
		if overwrite == nil {
			continue
		}
		out = append(out, events.PermissionOverwrite{
			ID:    overwrite.ID,
			Type:  int(overwrite.Type),
			Allow: overwrite.Allow,
			Deny:  overwrite.Deny,
		})
	}
	return out
}

func eventFromReaction(eventType events.Type, reaction *dgo.MessageReaction, member *dgo.Member) events.Event {
	if reaction == nil {
		return events.Event{Type: eventType}
	}
	event := events.Event{
		Type:      eventType,
		MessageID: reaction.MessageID,
		GuildID:   reaction.GuildID,
		ChannelID: reaction.ChannelID,
		UserID:    reaction.UserID,
		Reaction: &events.Reaction{
			EmojiName: reaction.Emoji.Name,
			EmojiID:   reaction.Emoji.ID,
		},
	}
	if member != nil {
		event.Member = memberFromDiscord(member)
	}
	return event
}

func eventFromMember(eventType events.Type, member *dgo.Member, guild *dgo.Guild, bot *dgo.User) events.Event {
	event := events.Event{Type: eventType}
	event.Member = memberFromDiscord(member)
	if event.Member != nil {
		event.GuildID = member.GuildID
		if guild != nil {
			event.GuildName = guild.Name
			event.GuildIconURL = guild.IconURL("")
		}
		if bot != nil {
			event.BotUserID = bot.ID
			event.BotAvatarURL = bot.AvatarURL("")
		}
		event.UserID = event.Member.UserID
	}
	return event
}

func guildFromState(session *dgo.Session, member *dgo.Member) *dgo.Guild {
	if session == nil || session.State == nil || member == nil || member.GuildID == "" {
		return nil
	}
	guild, err := session.State.Guild(member.GuildID)
	if err != nil || guild == nil {
		return nil
	}
	return guild
}

func botFromState(session *dgo.Session) *dgo.User {
	if session == nil || session.State == nil {
		return nil
	}
	return session.State.User
}

func eventFromVoiceState(voice *dgo.VoiceState, before *dgo.VoiceState, session *dgo.Session) events.Event {
	event := events.Event{Type: events.TypeVoiceState}
	if bot := botFromState(session); bot != nil {
		event.BotUserID = bot.ID
		event.BotAvatarURL = bot.AvatarURL("")
	}
	var member *dgo.Member
	if voice != nil {
		event.GuildID = voice.GuildID
		event.ChannelID = voice.ChannelID
		event.UserID = voice.UserID
		member = voice.Member
	}
	if before != nil {
		if event.GuildID == "" {
			event.GuildID = before.GuildID
		}
		if event.UserID == "" {
			event.UserID = before.UserID
		}
		if member == nil {
			member = before.Member
		}
	}
	event.Member = memberFromDiscord(member)
	if event.Member != nil {
		event.IsBot = event.Member.IsBot
		event.Username = event.Member.Username
		event.UserTag = event.Member.UserTag
		event.AvatarURL = event.Member.AvatarURL
		if event.UserID == "" {
			event.UserID = event.Member.UserID
		}
	}
	event.VoiceState = &events.VoiceState{
		UserID:    event.UserID,
		GuildID:   event.GuildID,
		ChannelID: event.ChannelID,
	}
	if voice != nil {
		event.VoiceState.ChannelName = channelNameFromState(session, voice.ChannelID)
	}
	if before != nil {
		event.VoiceState.BeforeChannel = before.ChannelID
		event.VoiceState.BeforeChannelName = channelNameFromState(session, before.ChannelID)
	}
	return event
}

func channelNameFromState(session *dgo.Session, channelID string) string {
	if session == nil || session.State == nil || channelID == "" {
		return ""
	}
	channel, err := session.State.Channel(channelID)
	if err != nil || channel == nil {
		return ""
	}
	return channel.Name
}

func memberFromDiscord(member *dgo.Member) *events.Member {
	if member == nil || member.User == nil {
		return nil
	}
	return &events.Member{
		UserID:           member.User.ID,
		Username:         member.User.Username,
		UserTag:          userTag(member.User),
		RoleIDs:          append([]string(nil), member.Roles...),
		JoinedAt:         member.JoinedAt,
		AccountCreatedAt: discordIDCreatedAt(member.User.ID),
		IsBot:            member.User.Bot,
		AvatarURL:        member.User.AvatarURL(""),
	}
}
