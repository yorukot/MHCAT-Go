package onboarding

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type WelcomeMessageDeliveryService struct {
	Repository ports.JoinMessageConfigReader
	Messages   ports.DiscordMessagePort
	Special    SpecialWelcomeConfig
}

type SpecialWelcomeConfig struct {
	GuildID          string
	BotID            string
	ChannelID        string
	ChatChannelID    string
	HelpChannelID    string
	BugChannelID     string
	SupportChannelID string
}

type WelcomeMemberEvent struct {
	GuildID       string
	GuildName     string
	GuildIconURL  string
	BotUserID     string
	BotAvatarURL  string
	UserID        string
	Username      string
	Discriminator string
	UserTag       string
	AvatarURL     string
	Now           time.Time
}

func (s WelcomeMessageDeliveryService) SendOnJoin(ctx context.Context, event WelcomeMemberEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.Messages == nil {
		return domain.ErrInvalidJoinMessageConfig
	}
	event.GuildID = strings.TrimSpace(event.GuildID)
	event.UserID = strings.TrimSpace(event.UserID)
	if event.GuildID == "" || event.UserID == "" {
		return nil
	}
	if s.Special.Matches(event) {
		_, err := s.Messages.SendMessage(ctx, s.Special.ChannelID, specialWelcomeMessage(event, s.Special))
		return err
	}
	if s.Repository == nil {
		return domain.ErrInvalidJoinMessageConfig
	}
	config, err := s.Repository.GetJoinMessageConfig(ctx, event.GuildID)
	if err != nil {
		if errors.Is(err, ports.ErrJoinMessageConfigMissing) {
			return nil
		}
		return err
	}
	if !config.Deliverable() {
		return nil
	}
	_, err = s.Messages.SendMessage(ctx, config.ChannelID, genericWelcomeMessage(config, event))
	return err
}

func (c SpecialWelcomeConfig) Matches(event WelcomeMemberEvent) bool {
	return strings.TrimSpace(c.GuildID) != "" &&
		strings.TrimSpace(c.BotID) != "" &&
		strings.TrimSpace(c.ChannelID) != "" &&
		strings.TrimSpace(c.ChatChannelID) != "" &&
		strings.TrimSpace(c.HelpChannelID) != "" &&
		strings.TrimSpace(c.BugChannelID) != "" &&
		strings.TrimSpace(c.SupportChannelID) != "" &&
		strings.TrimSpace(event.GuildID) == strings.TrimSpace(c.GuildID) &&
		strings.TrimSpace(event.BotUserID) == strings.TrimSpace(c.BotID)
}

func genericWelcomeMessage(config domain.JoinMessageConfig, event WelcomeMemberEvent) ports.OutboundMessage {
	return ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			AuthorName:    "🪂 歡迎加入 " + welcomeGuildName(event),
			AuthorIconURL: welcomeGuildIcon(event),
			Description:   replaceWelcomeMessagePlaceholders(config.MessageContent, event),
			Color:         welcomeMessageColor(config.Color),
			ThumbnailURL:  strings.TrimSpace(event.AvatarURL),
			ImageURL:      strings.TrimSpace(config.ImageURL),
			Timestamp:     welcomeTimestamp(event),
		}},
		AllowedMentions: ports.AllowedMentions{UserIDs: []string{strings.TrimSpace(event.UserID)}},
	}
}

func specialWelcomeMessage(event WelcomeMemberEvent, special SpecialWelcomeConfig) ports.OutboundMessage {
	return ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			AuthorName:    "🪂 歡迎加入 MHCAT!",
			AuthorIconURL: strings.TrimSpace(event.BotAvatarURL),
			AuthorURL:     "https://dsc.gg/MHCAT",
			Description:   specialWelcomeDescription(event, special),
			Color:         randomWelcomeMessageColor(),
			ThumbnailURL:  strings.TrimSpace(event.AvatarURL),
			ImageURL:      "https://i.imgur.com/cLCPRNq.png",
			Timestamp:     welcomeTimestamp(event),
		}},
		AllowedMentions: ports.AllowedMentions{},
	}
}

func replaceWelcomeMessagePlaceholders(value string, event WelcomeMemberEvent) string {
	memberName := event.Username
	if memberName == "" {
		memberName = usernameFromUserTag(event.UserTag)
	}
	if memberName == "" {
		memberName = strings.TrimSpace(event.UserID)
	}
	tag := "<@" + strings.TrimSpace(event.UserID) + ">"
	out := value
	out = legacyStringReplace(out, "(MEMBERNAME)", memberName)
	out = legacyStringReplace(out, "{MEMBERNAME}", memberName)
	out = legacyStringReplace(out, "{membername}", memberName)
	out = legacyStringReplace(out, "(TAG)", tag)
	out = legacyStringReplace(out, "{TAG}", tag)
	out = legacyStringReplace(out, "{tag}", tag)
	return out
}

func specialWelcomeDescription(event WelcomeMemberEvent, special SpecialWelcomeConfig) string {
	tag := strings.TrimSpace(event.UserTag)
	if event.Username != "" && event.Discriminator != "" {
		tag = event.Username + "#" + event.Discriminator
	}
	if tag == "" {
		tag = event.Username
	}
	if tag == "" {
		tag = strings.TrimSpace(event.UserID)
	}
	return "**<:welcome:978216428794679336> 歡迎 __" + tag + "__ 的加入!\n" +
		":speech_balloon: <#" + strings.TrimSpace(special.ChatChannelID) + ">想要聊天的話歡迎到這裡!\n" +
		"<:help:985948179709186058> <#" + strings.TrimSpace(special.HelpChannelID) + ">對指令有問題都可以到這邊問喔!\n" +
		"👾 <#" + strings.TrimSpace(special.BugChannelID) + ">有任何bug歡迎到這邊回報!\n    \n" +
		"如果有建議或試任何的問題或想法歡迎到\n" +
		"<#" + strings.TrimSpace(special.SupportChannelID) + ">開啟客服頻道**\n    \n" +
		"也祝你在這個伺服器內有個美好的回憶~\n        "
}

func welcomeGuildName(event WelcomeMemberEvent) string {
	name := strings.TrimSpace(event.GuildName)
	if name != "" {
		return name
	}
	return strings.TrimSpace(event.GuildID)
}

func welcomeGuildIcon(event WelcomeMemberEvent) string {
	if icon := strings.TrimSpace(event.GuildIconURL); icon != "" {
		return icon
	}
	return strings.TrimSpace(event.BotAvatarURL)
}

func welcomeTimestamp(event WelcomeMemberEvent) time.Time {
	if !event.Now.IsZero() {
		return event.Now
	}
	return time.Now()
}

func welcomeMessageColor(value string) int {
	if strings.TrimSpace(value) == "RANDOM" {
		return randomWelcomeMessageColor()
	}
	if parsed, ok := domain.ParseLegacyColorValue(value); ok {
		return parsed
	}
	return 0xED4245
}

func randomWelcomeMessageColor() int {
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(0x1000000))
	if err != nil {
		return 0x5865F2
	}
	return int(n.Int64())
}

func usernameFromUserTag(tag string) string {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return ""
	}
	if index := strings.Index(tag, "#"); index > 0 {
		return tag[:index]
	}
	return tag
}
