package notifications

import (
	"context"
	"errors"
	"sort"
	"sync"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
)

func TestDeliveryServiceReloadsAndSendsLegacyPayload(t *testing.T) {
	repo := newDeliveryRepository(domain.AutoNotificationSchedule{
		GuildID:   "guild-1",
		ID:        "schedule-1",
		Cron:      "0 9 * * 1",
		ChannelID: "channel-1",
		Message: domain.AutoNotificationMessage{
			Content:          "hello @everyone",
			EmbedTitle:       "Title",
			EmbedDescription: "Content",
			EmbedColor:       "#A6FFA6",
		},
	})
	messages := newAutoNotificationDeliverySideEffects()
	service := DeliveryService{Repository: repo, Messages: messages, Channels: messages}

	if err := service.Deliver(context.Background(), " guild-1 ", " schedule-1 "); err != nil {
		t.Fatalf("deliver: %v", err)
	}
	if len(messages.Sent) != 1 || messages.Sent[0].ChannelID != "channel-1" {
		t.Fatalf("sent = %#v", messages.Sent)
	}
	message := messages.Sent[0].Message
	if message.Content != "hello @everyone" || len(message.Embeds) != 1 || message.Embeds[0].Title != "Title" || message.Embeds[0].Description != "Content" || message.Embeds[0].Color != 0xA6FFA6 {
		t.Fatalf("message = %#v", message)
	}
	if !message.AllowedMentions.ParseEveryone || !message.AllowedMentions.ParseUsers || !message.AllowedMentions.ParseRoles {
		t.Fatalf("allowed mentions = %#v", message.AllowedMentions)
	}

	repo.delete("guild-1", "schedule-1")
	if err := service.Deliver(context.Background(), "guild-1", "schedule-1"); !errors.Is(err, ports.ErrAutoNotificationScheduleMissing) {
		t.Fatalf("expected missing schedule after delete, got %v", err)
	}
	if len(messages.Sent) != 1 {
		t.Fatalf("deleted schedule sent again: %#v", messages.Sent)
	}
}

func TestDeliveryServiceSupportsPersistedRandomColor(t *testing.T) {
	repo := newDeliveryRepository(domain.AutoNotificationSchedule{
		GuildID:   "guild-1",
		ID:        "schedule-1",
		Cron:      "0 9 * * 1",
		ChannelID: "channel-1",
		Message:   domain.AutoNotificationMessage{EmbedTitle: "Title", EmbedColor: "Random"},
	})
	messages := newAutoNotificationDeliverySideEffects()
	service := DeliveryService{Repository: repo, Messages: messages, Channels: messages, RandomColor: func() int { return 0x123456 }}
	if err := service.Deliver(context.Background(), "guild-1", "schedule-1"); err != nil {
		t.Fatalf("deliver: %v", err)
	}
	if len(messages.Sent) != 1 || len(messages.Sent[0].Message.Embeds) != 1 || messages.Sent[0].Message.Embeds[0].Color != 0x123456 {
		t.Fatalf("sent = %#v", messages.Sent)
	}
}

func TestDeliveryServicePreservesLegacyMessageWhitespace(t *testing.T) {
	repo := newDeliveryRepository(domain.AutoNotificationSchedule{
		GuildID:   "guild-1",
		ID:        "schedule-1",
		Cron:      "0 9 * * 1",
		ChannelID: "channel-1",
		Message: domain.AutoNotificationMessage{
			Content:          "  hello  ",
			EmbedTitle:       " ",
			EmbedDescription: "  Content  ",
		},
	})
	messages := newAutoNotificationDeliverySideEffects()
	service := DeliveryService{Repository: repo, Messages: messages, Channels: messages}
	if err := service.Deliver(context.Background(), "guild-1", "schedule-1"); err != nil {
		t.Fatalf("deliver: %v", err)
	}
	if len(messages.Sent) != 1 {
		t.Fatalf("sent = %#v", messages.Sent)
	}
	message := messages.Sent[0].Message
	if message.Content != "  hello  " || len(message.Embeds) != 1 || message.Embeds[0].Title != " " || message.Embeds[0].Description != "  Content  " {
		t.Fatalf("message = %#v", message)
	}
}

func TestDeliveryServiceRejectsChannelOutsideScheduleGuild(t *testing.T) {
	repo := newDeliveryRepository(domain.AutoNotificationSchedule{
		GuildID:   "guild-1",
		ID:        "schedule-1",
		Cron:      "0 9 * * 1",
		ChannelID: "channel-2",
		Message:   domain.AutoNotificationMessage{Content: "must not cross guilds"},
	})
	messages := newAutoNotificationDeliverySideEffects()
	service := DeliveryService{Repository: repo, Messages: messages, Channels: messages}

	if err := service.Deliver(context.Background(), "guild-1", "schedule-1"); !errors.Is(err, ports.ErrChannelNotFound) {
		t.Fatalf("deliver error = %v", err)
	}
	if len(messages.Sent) != 0 {
		t.Fatalf("sent = %#v", messages.Sent)
	}
}

func newAutoNotificationDeliverySideEffects() *fakediscord.SideEffects {
	sideEffects := fakediscord.NewSideEffects()
	sideEffects.Channels = []ports.ChannelRef{
		{GuildID: "guild-1", ChannelID: "channel-1"},
		{GuildID: "guild-2", ChannelID: "channel-2"},
	}
	return sideEffects
}

type deliveryRepository struct {
	mu        sync.Mutex
	schedules map[string]domain.AutoNotificationSchedule
}

func newDeliveryRepository(schedules ...domain.AutoNotificationSchedule) *deliveryRepository {
	repo := &deliveryRepository{schedules: map[string]domain.AutoNotificationSchedule{}}
	for _, schedule := range schedules {
		repo.set(schedule)
	}
	return repo
}

func (r *deliveryRepository) ListAutoNotificationDeliveries(ctx context.Context) ([]domain.AutoNotificationSchedule, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	keys := make([]string, 0, len(r.schedules))
	for key := range r.schedules {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]domain.AutoNotificationSchedule, 0, len(keys))
	for _, key := range keys {
		result = append(result, r.schedules[key])
	}
	return result, nil
}

func (r *deliveryRepository) GetAutoNotificationDelivery(ctx context.Context, guildID string, id string) (domain.AutoNotificationSchedule, error) {
	if err := ctx.Err(); err != nil {
		return domain.AutoNotificationSchedule{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	schedule, ok := r.schedules[autoNotificationDeliveryKey(guildID, id)]
	if !ok {
		return domain.AutoNotificationSchedule{}, ports.ErrAutoNotificationScheduleMissing
	}
	return schedule, nil
}

func (r *deliveryRepository) set(schedule domain.AutoNotificationSchedule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.schedules[autoNotificationDeliveryKey(schedule.GuildID, schedule.ID)] = schedule
}

func (r *deliveryRepository) delete(guildID string, id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.schedules, autoNotificationDeliveryKey(guildID, id))
}

var _ ports.AutoNotificationDeliveryRepository = (*deliveryRepository)(nil)
