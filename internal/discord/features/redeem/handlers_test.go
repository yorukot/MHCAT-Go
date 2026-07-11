package redeem

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/redeem"
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
	if len(responder.Defers) != 1 || !responder.Defers[0].Ephemeral {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if repo.Balances["guild-1"] != 5 {
		t.Fatalf("balances = %#v", repo.Balances)
	}
	if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	embed := responder.Edits[0].Embeds[0]
	if embed.Author == nil || embed.Author.Name != "成功兌換代碼!" || embed.Author.IconURL != successIconURL {
		t.Fatalf("embed author = %#v", embed.Author)
	}
	if embed.Footer == nil || embed.Footer.Text != "你可以使用/查看餘額進行查詢剩餘餘額" || embed.Color != redeemSuccessColor {
		t.Fatalf("embed = %#v", embed)
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
		want string
	}{
		{
			name: "missing",
			repo: fakemongo.NewRedeemRepository(),
			want: "找不到這個代碼!",
		},
		{
			name: "expired",
			repo: func() *fakemongo.RedeemRepository {
				repo := fakemongo.NewRedeemRepository()
				repo.Codes["abc"] = domain.RedeemCode{Code: "abc", Price: 1, CreatedAtMillis: float64(now.Add(-coreservice.LegacyCodeTTL - time.Millisecond).UnixMilli())}
				return repo
			}(),
			want: "這個代碼為防止遭人惡意使用",
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
			if len(responder.Edits) != 1 || len(responder.Edits[0].Embeds) != 1 {
				t.Fatalf("edits = %#v", responder.Edits)
			}
			if !strings.Contains(responder.Edits[0].Embeds[0].Title, tc.want) {
				t.Fatalf("title = %q", responder.Edits[0].Embeds[0].Title)
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
