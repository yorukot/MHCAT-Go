package domain_test

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestAutoNotificationDeliveryValidation(t *testing.T) {
	valid := domain.AutoNotificationSchedule{
		GuildID:   "guild-1",
		ID:        "schedule-1",
		Cron:      "0 9 * * 1",
		ChannelID: "channel-1",
		Message:   domain.AutoNotificationMessage{Content: "hello"},
	}
	if err := domain.ValidateAutoNotificationDelivery(valid); err != nil {
		t.Fatalf("valid delivery: %v", err)
	}
	valid.Pending = true
	if err := domain.ValidateAutoNotificationDelivery(valid); !errors.Is(err, domain.ErrInvalidAutoNotificationSchedule) {
		t.Fatalf("expected invalid pending delivery, got %v", err)
	}
}

func TestAutoNotificationColorAloneDoesNotCreateEmbed(t *testing.T) {
	message := domain.AutoNotificationMessage{Content: "hello", EmbedColor: "#123456"}
	if message.HasEmbed() {
		t.Fatal("color without title or description should preserve the legacy plain-message shape")
	}
}

func TestAutoNotificationMessagePreservesLegacyWhitespace(t *testing.T) {
	plain := domain.AutoNotificationMessage{Content: "   "}.Normalized()
	if plain.Content != "   " || plain.Empty() || plain.HasEmbed() {
		t.Fatalf("plain message = %#v", plain)
	}
	embed := domain.AutoNotificationMessage{EmbedTitle: " ", EmbedDescription: "  content  "}.Normalized()
	if embed.EmbedTitle != " " || embed.EmbedDescription != "  content  " || embed.Empty() || !embed.HasEmbed() {
		t.Fatalf("embed message = %#v", embed)
	}
}
