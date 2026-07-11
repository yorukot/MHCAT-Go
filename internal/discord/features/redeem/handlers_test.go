package redeem

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/redeem"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func TestHandlerRedeemsCodeAndRendersLegacySuccess(t *testing.T) {
	now := time.UnixMilli(1700000000000)
	repo := fakemongo.NewRedeemRepository()
	repo.Codes[" abc "] = domain.RedeemCode{Code: " abc ", Price: 5, CreatedAtMillis: float64(now.UnixMilli())}
	usage := &fakeusage.Tracker{}
	module := NewModule(repo, fixedClock{now: now}, usage)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(CommandName, "", map[string]string{optionCode: " abc "})

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !reflect.DeepEqual(responder.Defers, []responses.DeferOptions{{Ephemeral: true}}) {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if repo.Balances["guild-1"] != 5 {
		t.Fatalf("balances = %#v", repo.Balances)
	}
	want := responses.Message{
		Embeds: []responses.Embed{{
			Author: &responses.EmbedAuthor{
				Name:    "成功兌換代碼!",
				IconURL: "https://media.discordapp.net/attachments/991337796960784424/1078883215462383697/success.gif",
			},
			Footer: &responses.EmbedFooter{Text: "你可以使用/查看餘額進行查詢剩餘餘額"},
			Color:  0x57F287,
		}},
		Ephemeral:       true,
		AllowedMentions: &responses.AllowedMentions{},
	}
	if len(responder.Edits) != 1 || !reflect.DeepEqual(responder.Edits[0], want) {
		t.Fatalf("edits = %#v, want %#v", responder.Edits, want)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != CommandName || usage.Events[0].Feature != "redeem" {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestHandlerUsesLegacyErrors(t *testing.T) {
	now := time.UnixMilli(1700000000000)
	tests := []struct {
		name string
		repo *fakemongo.RedeemRepository
		want responses.Message
	}{
		{
			name: "missing",
			repo: fakemongo.NewRedeemRepository(),
			want: redeemErrorMessage("找不到這個代碼!"),
		},
		{
			name: "expired",
			repo: func() *fakemongo.RedeemRepository {
				repo := fakemongo.NewRedeemRepository()
				repo.Codes["abc"] = domain.RedeemCode{Code: "abc", Price: 1, CreatedAtMillis: float64(now.Add(-coreservice.LegacyCodeTTL - time.Millisecond).UnixMilli())}
				return repo
			}(),
			want: redeemErrorMessage("這個代碼為防止遭人惡意使用，已過期，如果你是代碼擁有者，請前往支援伺服器開啟客服頻道!"),
		},
		{
			name: "backend failure",
			repo: func() *fakemongo.RedeemRepository {
				repo := fakemongo.NewRedeemRepository()
				repo.Err = errors.New("mongo credential secret")
				return repo
			}(),
			want: redeemErrorMessage("很抱歉，出現了未知的錯誤，請重試!"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			module := NewModule(tc.repo, fixedClock{now: now}, nil)
			responder := fakediscord.NewResponder()
			interaction := fakediscord.SlashInteractionWithOptions(CommandName, "", map[string]string{optionCode: "abc"})
			if err := module.Handler()(context.Background(), interaction, responder); err != nil {
				t.Fatalf("handler: %v", err)
			}
			if !reflect.DeepEqual(responder.Defers, []responses.DeferOptions{{Ephemeral: true}}) {
				t.Fatalf("defers = %#v", responder.Defers)
			}
			if len(responder.Edits) != 1 || !reflect.DeepEqual(responder.Edits[0], tc.want) {
				t.Fatalf("edits = %#v, want %#v", responder.Edits, tc.want)
			}
			if len(tc.repo.Balances) != 0 {
				t.Fatalf("unexpected balance update = %#v", tc.repo.Balances)
			}
		})
	}
}

func TestErrorTextFallsBackForUnknownErrors(t *testing.T) {
	if got := redeemErrorText(errors.New("boom")); !strings.Contains(got, "未知的錯誤") {
		t.Fatalf("unexpected fallback = %q", got)
	}
	if got := redeemErrorText(ports.ErrRedeemCodeNotFound); got != "找不到這個代碼!" {
		t.Fatalf("not found = %q", got)
	}
}
