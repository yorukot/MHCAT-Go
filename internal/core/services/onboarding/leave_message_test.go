package onboarding

import (
	"context"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
)

func TestLeaveMessageServicePrepareNormalizesChannel(t *testing.T) {
	repo := &fakeLeaveMessageRepo{}
	service := LeaveMessageService{Repository: repo}
	config, err := service.Prepare(context.Background(), " guild ", " channel ")
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if config.GuildID != "guild" || config.ChannelID != "channel" {
		t.Fatalf("config = %#v", config)
	}
	if repo.preparedGuild != "guild" || repo.preparedChannel != "channel" {
		t.Fatalf("repo fields = %q/%q", repo.preparedGuild, repo.preparedChannel)
	}
}

func TestLeaveMessageServiceSaveValidatesColor(t *testing.T) {
	repo := &fakeLeaveMessageRepo{prepared: domain.LeaveMessageConfig{GuildID: "guild", ChannelID: "channel"}}
	service := LeaveMessageService{Repository: repo}
	err := service.Save(context.Background(), domain.LeaveMessageConfig{
		GuildID:        " guild ",
		MessageContent: "bye",
		Title:          "bye title",
		Color:          " not-a-color ",
	})
	if err == nil {
		t.Fatal("expected invalid color")
	}
	if repo.saved.GuildID != "" {
		t.Fatalf("should not save invalid config: %#v", repo.saved)
	}
}

func TestLeaveMessageServiceSave(t *testing.T) {
	repo := &fakeLeaveMessageRepo{prepared: domain.LeaveMessageConfig{GuildID: "guild", ChannelID: "channel"}}
	service := LeaveMessageService{Repository: repo}
	err := service.Save(context.Background(), domain.LeaveMessageConfig{
		GuildID:        " guild ",
		MessageContent: "bye",
		Title:          "bye title",
		Color:          "#df1f2f",
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if repo.saved.GuildID != "guild" || repo.saved.Color != "#df1f2f" {
		t.Fatalf("saved = %#v", repo.saved)
	}
}

func TestLeaveMessageServiceRejectsPaddedColor(t *testing.T) {
	repo := &fakeLeaveMessageRepo{prepared: domain.LeaveMessageConfig{GuildID: "guild", ChannelID: "channel"}}
	service := LeaveMessageService{Repository: repo}
	for _, color := range []string{" #df1f2f ", " Random "} {
		err := service.Save(context.Background(), domain.LeaveMessageConfig{
			GuildID:        "guild",
			MessageContent: "bye",
			Title:          "bye title",
			Color:          color,
		})
		if err == nil {
			t.Fatalf("expected padded color %q to fail", color)
		}
	}
}

func TestLeaveMessageServicePreservesAllSpaceText(t *testing.T) {
	repo := &fakeLeaveMessageRepo{prepared: domain.LeaveMessageConfig{GuildID: "guild", ChannelID: "channel"}}
	service := LeaveMessageService{Repository: repo}
	if err := service.Save(context.Background(), domain.LeaveMessageConfig{
		GuildID:        "guild",
		MessageContent: "   ",
		Title:          "  ",
		Color:          "#df1f2f",
	}); err != nil {
		t.Fatalf("save: %v", err)
	}
	if repo.saved.MessageContent != "   " || repo.saved.Title != "  " {
		t.Fatalf("saved = %#v", repo.saved)
	}
}

func TestLeaveMessageDeliveryServiceSendsLegacyLeaveEmbed(t *testing.T) {
	now := time.Date(2026, 7, 4, 1, 2, 3, 0, time.UTC)
	repo := &fakeLeaveMessageRepo{prepared: domain.LeaveMessageConfig{
		GuildID:        "guild",
		ChannelID:      "channel",
		Title:          "Bye (MEMBERNAME)",
		MessageContent: "Goodbye {MEMBERNAME} / (ID) / {ID}",
		Color:          "#df1f2f",
	}}
	messages := &fakeLeaveMessageSender{}
	channels := fakediscord.NewSideEffects()
	cacheLegacyDeliveryChannel(channels, "guild", "channel")
	service := LeaveMessageDeliveryService{Repository: repo, Messages: messages, Channels: channels}

	err := service.SendOnLeave(context.Background(), LeaveMemberEvent{
		GuildID:   " guild ",
		UserID:    "user-1",
		Username:  "Tester",
		UserTag:   "Tester#0001",
		AvatarURL: "https://cdn.example/avatar.png",
		Now:       now,
	})
	if err != nil {
		t.Fatalf("send leave message: %v", err)
	}
	if len(messages.sent) != 1 {
		t.Fatalf("sent = %#v", messages.sent)
	}
	msg := messages.sent[0]
	if msg.channelID != "channel" || len(msg.message.Embeds) != 1 {
		t.Fatalf("message = %#v", msg)
	}
	embed := msg.message.Embeds[0]
	if embed.Title != "Bye (MEMBERNAME)" || embed.Description != "Goodbye Tester / user-1 / user-1" {
		t.Fatalf("embed text = %#v", embed)
	}
	if embed.Color != 0xDF1F2F || embed.ThumbnailURL != "https://cdn.example/avatar.png" || !embed.Timestamp.Equal(now) {
		t.Fatalf("embed shape = %#v", embed)
	}
}

func TestLeaveMessageDeliveryServiceNoopsForMissingOrIncompleteConfig(t *testing.T) {
	for name, config := range map[string]domain.LeaveMessageConfig{
		"missing":    {},
		"no-channel": {GuildID: "guild", Title: "Bye", MessageContent: "Bye", Color: "#df1f2f"},
		"no-content": {GuildID: "guild", ChannelID: "channel", Title: "Bye", Color: "#df1f2f"},
	} {
		t.Run(name, func(t *testing.T) {
			repo := &fakeLeaveMessageRepo{prepared: config}
			messages := &fakeLeaveMessageSender{}
			service := LeaveMessageDeliveryService{Repository: repo, Messages: messages, Channels: fakediscord.NewSideEffects()}
			if err := service.SendOnLeave(context.Background(), LeaveMemberEvent{GuildID: "guild", UserID: "user"}); err != nil {
				t.Fatalf("send: %v", err)
			}
			if len(messages.sent) != 0 {
				t.Fatalf("sent = %#v", messages.sent)
			}
		})
	}
}

func TestLeaveMessageDeliveryPreservesAllSpaceText(t *testing.T) {
	now := time.Unix(2_000_000, 0)
	repo := &fakeLeaveMessageRepo{prepared: domain.LeaveMessageConfig{
		GuildID:        "guild",
		ChannelID:      "channel",
		Title:          "  ",
		MessageContent: "   ",
		Color:          "#df1f2f",
	}}
	messages := &fakeLeaveMessageSender{}
	channels := fakediscord.NewSideEffects()
	cacheLegacyDeliveryChannel(channels, "guild", "channel")
	service := LeaveMessageDeliveryService{Repository: repo, Messages: messages, Channels: channels}

	if err := service.SendOnLeave(context.Background(), LeaveMemberEvent{GuildID: "guild", UserID: "user", Now: now}); err != nil {
		t.Fatalf("send: %v", err)
	}
	if len(messages.sent) != 1 || messages.sent[0].message.Embeds[0].Title != "  " || messages.sent[0].message.Embeds[0].Description != "   " {
		t.Fatalf("sent = %#v", messages.sent)
	}
}

func TestLeaveMessageDeliverySkipsUncachedLegacyChannel(t *testing.T) {
	repo := &fakeLeaveMessageRepo{prepared: domain.LeaveMessageConfig{
		GuildID: "guild", ChannelID: "missing", Title: "Bye", MessageContent: "Bye", Color: "Red",
	}}
	messages := &fakeLeaveMessageSender{}
	service := LeaveMessageDeliveryService{Repository: repo, Messages: messages, Channels: fakediscord.NewSideEffects()}
	if err := service.SendOnLeave(context.Background(), LeaveMemberEvent{GuildID: "guild", UserID: "user"}); err != nil {
		t.Fatalf("send: %v", err)
	}
	if len(messages.sent) != 0 {
		t.Fatalf("sent = %#v", messages.sent)
	}
}

func TestLeaveMessagePlaceholdersPreserveJavaScriptReplacementTokens(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     string
	}{
		{name: "dollar", username: "$$", want: "pre$post"},
		{name: "matched text", username: "$&", want: "pre{MEMBERNAME}post"},
		{name: "prefix", username: "$`", want: "preprepost"},
		{name: "suffix", username: "$'", want: "prepostpost"},
		{name: "raw whitespace", username: "  Yoru  ", want: "pre  Yoru  post"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := replaceLeaveMessageDescriptionPlaceholders("pre{MEMBERNAME}post", LeaveMemberEvent{UserID: "user", Username: tc.username})
			if got != tc.want {
				t.Fatalf("description = %q, want %q", got, tc.want)
			}
		})
	}
}

type fakeLeaveMessageRepo struct {
	preparedGuild   string
	preparedChannel string
	prepared        domain.LeaveMessageConfig
	saved           domain.LeaveMessageConfig
}

func (r *fakeLeaveMessageRepo) PrepareLeaveMessageConfig(ctx context.Context, guildID string, channelID string) (domain.LeaveMessageConfig, error) {
	r.preparedGuild = guildID
	r.preparedChannel = channelID
	if r.prepared.GuildID == "" {
		r.prepared = domain.LeaveMessageConfig{GuildID: guildID}
	}
	r.prepared.ChannelID = channelID
	return r.prepared, ctx.Err()
}

func (r *fakeLeaveMessageRepo) SaveLeaveMessageContent(ctx context.Context, config domain.LeaveMessageConfig) error {
	r.saved = config
	return ctx.Err()
}

func (r *fakeLeaveMessageRepo) GetLeaveMessageConfig(ctx context.Context, guildID string) (domain.LeaveMessageConfig, error) {
	if r.prepared.GuildID == "" {
		return domain.LeaveMessageConfig{}, ports.ErrLeaveMessageConfigMissing
	}
	return r.prepared, ctx.Err()
}

type fakeLeaveMessageSender struct {
	sent []struct {
		channelID string
		message   ports.OutboundMessage
	}
}

func (s *fakeLeaveMessageSender) SendMessage(ctx context.Context, channelID string, msg ports.OutboundMessage) (ports.MessageRef, error) {
	s.sent = append(s.sent, struct {
		channelID string
		message   ports.OutboundMessage
	}{channelID: channelID, message: msg})
	return ports.MessageRef{ChannelID: channelID, MessageID: "message"}, ctx.Err()
}

func (s *fakeLeaveMessageSender) EditMessage(ctx context.Context, ref ports.MessageRef, msg ports.OutboundMessage) error {
	return ctx.Err()
}

func (s *fakeLeaveMessageSender) DeleteMessage(ctx context.Context, ref ports.MessageRef) error {
	return ctx.Err()
}
