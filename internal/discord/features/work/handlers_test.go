package work

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestAddWorkSubcommandUsesLegacyDashboardRedirect(t *testing.T) {
	usage := &fakeusage.Tracker{}
	module := NewModule(usage)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(CommandName, subcommandAddWork, nil)
	interaction.Actor.GuildID = "guild-123"

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handle work dashboard redirect: %v", err)
	}
	if len(responder.Replies) != 1 {
		t.Fatalf("expected one reply, got %#v", responder.Replies)
	}
	reply := responder.Replies[0]
	if len(reply.Embeds) != 1 || reply.Embeds[0].Title != dashboardTitle || reply.Embeds[0].Color != dashboardColor {
		t.Fatalf("unexpected embed: %#v", reply.Embeds)
	}
	if len(reply.Components) != 1 || len(reply.Components[0].Components) != 1 {
		t.Fatalf("expected one link button row, got %#v", reply.Components)
	}
	button := reply.Components[0].Components[0]
	if button.Label != dashboardLabel || button.Emoji != dashboardEmoji || button.URL != "https://mhcat.yorukot.me//guilds/guild-123/work" {
		t.Fatalf("unexpected button: %#v", button)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != CommandName || usage.Events[0].Feature != "work" {
		t.Fatalf("unexpected usage events: %#v", usage.Events)
	}
}

func TestUnimplementedWorkSubcommandReturnsSafeEphemeralMessage(t *testing.T) {
	module := NewModule(nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(CommandName, "打工介面", nil)
	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handle unimplemented work subcommand: %v", err)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral {
		t.Fatalf("expected one ephemeral reply, got %#v", responder.Replies)
	}
	if len(responder.Replies[0].Embeds) != 1 || !strings.Contains(responder.Replies[0].Embeds[0].Title, "打工介面") {
		t.Fatalf("unexpected unimplemented message: %#v", responder.Replies[0])
	}
}

func TestWorkInterfaceRendersLegacyListReadOnly(t *testing.T) {
	repo := fakemongo.NewWorkInterfaceRepository()
	repo.PutConfig(domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 20})
	item := domain.WorkItem{GuildID: "guild-1", Name: "礦坑", DurationSec: 3600, EnergyCost: 3, CoinReward: 88}
	repo.PutItems("guild-1", item)
	repo.PutUser(domain.WorkUserState{GuildID: "guild-1", UserID: "user-1", State: domain.WorkIdleState, Energy: 9, Initialized: true})
	module := NewModuleWithRepository(repo, nil, nil)
	module.color = func() int { return 0x123456 }
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(CommandName, subcommandWorkInterface, nil)
	interaction.ChannelName = "測試伺服器"
	interaction.Actor.UserTag = "tester#0001"

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handle work interface: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 {
		t.Fatalf("expected defer/edit, defers=%#v edits=%#v replies=%#v", responder.Defers, responder.Edits, responder.Replies)
	}
	edit := responder.Edits[0]
	if edit.Embeds[0].Color != 0x123456 {
		t.Fatalf("list color = %#x", edit.Embeds[0].Color)
	}
	if len(edit.Embeds) != 1 || edit.Embeds[0].Title != "<:list:992002476360343602> 以下是測試伺服器的打工簡章" {
		t.Fatalf("unexpected embed title: %#v", edit.Embeds)
	}
	if !strings.Contains(edit.Embeds[0].Description, "**剩餘體力:** `9 \\ 20`") {
		t.Fatalf("unexpected description: %q", edit.Embeds[0].Description)
	}
	if len(edit.Embeds[0].Fields) != 1 || !strings.Contains(edit.Embeds[0].Fields[0].Value, "`60分(1小時)`") {
		t.Fatalf("unexpected fields: %#v", edit.Embeds[0].Fields)
	}
	if len(edit.Components) != 1 || len(edit.Components[0].Components) != 1 {
		t.Fatalf("unexpected components: %#v", edit.Components)
	}
	button := edit.Components[0].Components[0]
	if button.Label != "礦坑" || !strings.HasPrefix(button.CustomID, "mhcat:v1:work:detail:job=") {
		t.Fatalf("unexpected work button: %#v", button)
	}
	if !strings.Contains(button.CustomID, "user=user-1") {
		t.Fatalf("expected requester-bound custom id, got %q", button.CustomID)
	}
	if len(button.CustomID) > customid.MaxCustomIDLength {
		t.Fatalf("custom id too long: %d", len(button.CustomID))
	}
}

func TestWorkInterfaceRendersPreservedScalarText(t *testing.T) {
	view := domain.WorkInterfaceView{
		Config: domain.WorkConfig{MaxEnergy: 20, MaxEnergyText: "20.5"},
		User: domain.WorkUserState{
			State:      domain.WorkIdleState,
			Energy:     9,
			EnergyText: "9.5",
		},
	}
	if got := workInterfaceDescription(view); !strings.Contains(got, "`9.5 \\ 20.5`") {
		t.Fatalf("interface description = %q", got)
	}

	item := domain.WorkItem{
		Name:           "礦坑",
		DurationText:   "90.5",
		EnergyCostText: "3.5",
		CoinRewardText: "88.5",
	}
	field := workItemFields([]domain.WorkItem{item})[0].Value
	for _, want := range []string{"`3.5`", "`1.5083333333333333分(0.025138888888888888小時)`", "`88.5`"} {
		if !strings.Contains(field, want) {
			t.Fatalf("work list field %q does not contain %q", field, want)
		}
	}
	detail := workDetailMessage(view, item, false, 0).Embeds[0].Description
	for _, want := range []string{"`88.5 個代幣`", "`3.5`"} {
		if !strings.Contains(detail, want) {
			t.Fatalf("work detail %q does not contain %q", detail, want)
		}
	}
}

func TestWorkListRendersLegacyDurationCoercion(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "null", raw: "null", want: "`0分(0小時)`"},
		{name: "infinity", raw: "Infinity", want: "`Infinity分(Infinity小時)`"},
		{name: "malformed", raw: "undefined", want: "`NaN分(NaN小時)`"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			field := workItemFields([]domain.WorkItem{{Name: "礦坑", DurationText: test.raw}})[0].Value
			if !strings.Contains(field, test.want) {
				t.Fatalf("work list field = %q, want %q", field, test.want)
			}
		})
	}
}

func TestWorkInterfaceShowsCaptchaModalWhenEnabled(t *testing.T) {
	repo := fakemongo.NewWorkInterfaceRepository()
	repo.PutConfig(domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 20, Captcha: true})
	module := NewModuleWithRepository(repo, nil, nil)
	module.captcha = func() (int, int) { return 2, 3 }
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(CommandName, subcommandWorkInterface, nil)

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handle work captcha prompt: %v", err)
	}
	if len(responder.Modals) != 1 || responder.Modals[0].Title != "認證你不是機器人!" {
		t.Fatalf("modals = %#v", responder.Modals)
	}
	if responder.Modals[0].Rows[0].Inputs[0].Label != "請計算2 + 3" {
		t.Fatalf("modal input = %#v", responder.Modals[0].Rows[0].Inputs[0])
	}
	if !strings.Contains(responder.Modals[0].CustomID, "sum=5") {
		t.Fatalf("modal custom id = %q", responder.Modals[0].CustomID)
	}
}

func TestWorkSettingsRequiresManageMessages(t *testing.T) {
	repo := fakemongo.NewWorkInterfaceRepository()
	module := NewModuleWithAdminRepository(repo, nil, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(CommandName, subcommandWorkSettings, map[string]string{
		"每天可獲得多少精力": "5",
		"精力上限為多少":   "20",
	})

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("settings handler: %v", err)
	}
	if len(responder.Defers) != 1 || len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "訊息管理") {
		t.Fatalf("expected permission edit, defers=%#v edits=%#v", responder.Defers, responder.Edits)
	}
}

func TestWorkSettingsSavesLegacyConfig(t *testing.T) {
	repo := fakemongo.NewWorkInterfaceRepository()
	module := NewModuleWithAdminRepository(repo, nil, nil)
	responder := fakediscord.NewResponder()
	interaction := workAdminSlash(subcommandWorkSettings, map[string]string{
		"每天可獲得多少精力": "5",
		"精力上限為多少":   "20",
		"是否需要驗證":    "true",
	})

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("settings handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<:working:1048617967799242772> 打工系統" {
		t.Fatalf("settings edit = %#v", responder.Edits)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Description, "**成功設定打工系統!**") || !strings.Contains(responder.Edits[0].Embeds[0].Description, "`true`") {
		t.Fatalf("settings description = %q", responder.Edits[0].Embeds[0].Description)
	}
	config, err := repo.GetWorkConfig(context.Background(), "guild-1")
	if err != nil || config.DailyEnergy != 5 || config.MaxEnergy != 20 || !config.Captcha {
		t.Fatalf("saved config = %#v, %v", config, err)
	}
}

func TestDeleteWorkUsesLegacyMessages(t *testing.T) {
	repo := fakemongo.NewWorkInterfaceRepository()
	repo.PutConfig(domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 20})
	repo.PutItems("guild-1", domain.WorkItem{GuildID: "guild-1", Name: "礦坑"})
	module := NewModuleWithAdminRepository(repo, nil, nil)
	responder := fakediscord.NewResponder()
	interaction := workAdminSlash(subcommandDeleteWork, map[string]string{"打工地點名稱": "礦坑"})

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("delete handler: %v", err)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<:working:1048617967799242772> 打工事項" {
		t.Fatalf("delete edit = %#v", responder.Edits)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Description, "**成功刪除打工事項**:`礦坑`") {
		t.Fatalf("delete description = %q", responder.Edits[0].Embeds[0].Description)
	}
	if _, err := repo.ListWorkItems(context.Background(), "guild-1"); err == nil {
		t.Fatal("expected work item to be deleted")
	}
}

func TestDeleteWorkMissingItemUsesLegacyError(t *testing.T) {
	repo := fakemongo.NewWorkInterfaceRepository()
	repo.PutConfig(domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 20})
	module := NewModuleWithAdminRepository(repo, nil, nil)
	responder := fakediscord.NewResponder()
	interaction := workAdminSlash(subcommandDeleteWork, map[string]string{"打工地點名稱": "不存在"})

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("delete handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "找不到這個名字的資料") {
		t.Fatalf("missing item edit = %#v", responder.Edits)
	}
}

func TestAddUserEnergyClampsAndCreatesTargetUser(t *testing.T) {
	repo := fakemongo.NewWorkInterfaceRepository()
	repo.PutConfig(domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 20})
	module := NewModuleWithAdminRepository(repo, nil, nil)
	responder := fakediscord.NewResponder()
	interaction := workAdminSlash(subcommandAddUserEnergy, map[string]string{"使用者": "target-user", "要給多少精力": "5"})
	interaction.Actor.UserID = "admin-user"

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("grant user handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "<@target-user>") {
		t.Fatalf("grant user edit = %#v", responder.Edits)
	}
	target, err := repo.GetWorkUser(context.Background(), "guild-1", "target-user")
	if err != nil || target.UserID != "target-user" || target.Energy != 20 {
		t.Fatalf("target user = %#v, %v", target, err)
	}
	if _, err := repo.GetWorkUser(context.Background(), "guild-1", "admin-user"); err == nil {
		t.Fatal("legacy bug regression: admin user row should not be created for target grant")
	}
}

func TestAddAllEnergyClampsExistingUsersOnly(t *testing.T) {
	repo := fakemongo.NewWorkInterfaceRepository()
	repo.PutConfig(domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 20})
	repo.PutUser(domain.WorkUserState{GuildID: "guild-1", UserID: "user-1", State: domain.WorkIdleState, Energy: 19, Initialized: true})
	repo.PutUser(domain.WorkUserState{GuildID: "guild-1", UserID: "user-2", State: domain.WorkIdleState, Energy: 10, Initialized: true})
	repo.PutUser(domain.WorkUserState{GuildID: "other", UserID: "user-3", State: domain.WorkIdleState, Energy: 1, Initialized: true})
	module := NewModuleWithAdminRepository(repo, nil, nil)
	responder := fakediscord.NewResponder()
	interaction := workAdminSlash(subcommandAddAllEnergy, map[string]string{"要給多少精力": "5"})

	if err := module.Handler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("grant all handler: %v", err)
	}
	if len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Description, "所有已建檔的使用者") {
		t.Fatalf("grant all edit = %#v", responder.Edits)
	}
	user1, _ := repo.GetWorkUser(context.Background(), "guild-1", "user-1")
	user2, _ := repo.GetWorkUser(context.Background(), "guild-1", "user-2")
	other, _ := repo.GetWorkUser(context.Background(), "other", "user-3")
	if user1.Energy != 20 || user2.Energy != 15 || other.Energy != 1 {
		t.Fatalf("unexpected energies user1=%#v user2=%#v other=%#v", user1, user2, other)
	}
}

func TestWorkCaptchaModalWrongAnswerUsesLegacyErrorContent(t *testing.T) {
	repo := fakemongo.NewWorkInterfaceRepository()
	repo.PutConfig(domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 20})
	module := NewModuleWithRepository(repo, nil, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ModalInteraction(interactions.ModalKey{})
	interaction.CustomID = "mhcat:v1:work:captcha:sum=5"
	interaction.ModalFields = []customid.ModalField{{CustomID: "captcha", Value: "4"}}

	if err := module.CaptchaHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("captcha handler: %v", err)
	}
	if len(responder.Replies) != 1 || !strings.Contains(responder.Replies[0].Content, "驗證碼錯誤") {
		t.Fatalf("replies = %#v", responder.Replies)
	}
}

func TestWorkCaptchaModalUsesLegacyNumberCoercion(t *testing.T) {
	for _, answer := range []string{"5", "05", " 5 ", "5e0", "0x5", "0b101", "0o5"} {
		t.Run(answer, func(t *testing.T) {
			repo := fakemongo.NewWorkInterfaceRepository()
			repo.PutConfig(domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 20})
			repo.PutItems("guild-1", domain.WorkItem{GuildID: "guild-1", Name: "礦坑", DurationSec: 60, EnergyCost: 1, CoinReward: 1})
			module := NewModuleWithRepository(repo, nil, nil)
			responder := fakediscord.NewResponder()
			interaction := fakediscord.ModalInteraction(interactions.ModalKey{})
			interaction.CustomID = "mhcat:v1:work:captcha:sum=5"
			interaction.ModalFields = []customid.ModalField{{CustomID: "captcha", Value: answer}}

			if err := module.CaptchaHandler()(context.Background(), interaction, responder); err != nil {
				t.Fatalf("captcha handler: %v", err)
			}
			if len(responder.Defers) != 1 || len(responder.Edits) != 1 || !strings.Contains(responder.Edits[0].Embeds[0].Title, "打工簡章") {
				t.Fatalf("response = defers %#v edits %#v", responder.Defers, responder.Edits)
			}
		})
	}
}

func TestWorkDetailRendersLegacyDetailWithDisabledConfirmWhenReadOnly(t *testing.T) {
	repo := fakemongo.NewWorkInterfaceRepository()
	repo.PutConfig(domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 20})
	item := domain.WorkItem{GuildID: "guild-1", Name: "礦坑", DurationSec: 3600, EnergyCost: 3, CoinReward: 88}
	repo.PutItems("guild-1", item)
	repo.PutUser(domain.WorkUserState{GuildID: "guild-1", UserID: "user-1", State: domain.WorkIdleState, Energy: 9, Initialized: true})
	module := NewModuleWithRepository(repo, nil, nil)
	module.color = func() int { return 0x654321 }
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID(workDetailCustomID(item))

	if err := module.DetailHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("detail handler: %v", err)
	}
	if len(responder.Updates) != 1 || len(responder.Updates[0].Embeds) != 1 {
		t.Fatalf("updates = %#v", responder.Updates)
	}
	update := responder.Updates[0]
	if update.Embeds[0].Color != 0x654321 {
		t.Fatalf("detail color = %#x", update.Embeds[0].Color)
	}
	if update.Embeds[0].Title != "<:creativeteaching:986060052949524600> 以下是礦坑打工的詳細資料" {
		t.Fatalf("detail embed = %#v", update.Embeds[0])
	}
	if len(update.Components) != 1 || len(update.Components[0].Components) != 1 {
		t.Fatalf("expected confirm button, got %#v", update.Components)
	}
	button := update.Components[0].Components[0]
	if !button.Disabled || button.Label != "確認打工" || button.Style != "success" || !strings.HasPrefix(button.CustomID, "mhcat:v1:work:start:job=") {
		t.Fatalf("unexpected confirm button: %#v", button)
	}
}

func TestWorkDetailRendersActiveConfirmWithStartRepository(t *testing.T) {
	repo, item := workTestRepository()
	module := NewModuleWithStartRepository(repo, nil, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID(workDetailCustomID(item, "user-1"))

	if err := module.DetailHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("detail handler: %v", err)
	}
	if len(responder.Updates) != 1 || len(responder.Updates[0].Components) != 1 {
		t.Fatalf("updates = %#v", responder.Updates)
	}
	button := responder.Updates[0].Components[0].Components[0]
	if button.Disabled || button.Label != "確認打工" || button.Style != "success" {
		t.Fatalf("expected active confirm button, got %#v", button)
	}
}

func TestWorkDetailRejectsDifferentRequester(t *testing.T) {
	repo, item := workTestRepository()
	module := NewModuleWithRepository(repo, nil, nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID(workDetailCustomID(item, "user-1"))
	interaction.Actor.UserID = "user-2"

	if err := module.DetailHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("detail handler: %v", err)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral {
		t.Fatalf("expected one ephemeral reply, got %#v", responder.Replies)
	}
	if !strings.Contains(responder.Replies[0].Embeds[0].Title, "你不是查詢者無法使用") {
		t.Fatalf("unexpected unauthorized message: %#v", responder.Replies[0])
	}
}

func TestWorkStartSuccessUsesLegacySuccessMessageAndAtomicFakeUpdate(t *testing.T) {
	repo, item := workTestRepository()
	module := NewModuleWithStartRepository(repo, nil, fixedClock{now: time.Unix(100, 0)})
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID(workStartCustomID(item, "user-1"))

	if err := module.StartHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("start handler: %v", err)
	}
	if len(responder.Updates) != 1 || len(responder.Updates[0].Embeds) != 1 {
		t.Fatalf("updates = %#v", responder.Updates)
	}
	embed := responder.Updates[0].Embeds[0]
	if embed.Title != "<:working:1048617967799242772> 成功取得該工作!" || embed.Color != workSuccessColor {
		t.Fatalf("unexpected success embed: %#v", embed)
	}
	if !strings.Contains(embed.Description, "**你已經成功取得**`礦坑`**的工作**") || !strings.Contains(embed.Description, "<t:3700:R>") {
		t.Fatalf("unexpected success description: %q", embed.Description)
	}
	user, err := repo.GetWorkUser(context.Background(), "guild-1", "user-1")
	if err != nil {
		t.Fatalf("get updated user: %v", err)
	}
	if user.State != "礦坑" || user.Energy != 6 || user.GetCoin != 88 || user.EndTimeUnix != 3700 {
		t.Fatalf("unexpected updated user: %#v", user)
	}
}

func TestWorkStartInsufficientEnergyUsesLegacyError(t *testing.T) {
	repo, item := workTestRepository()
	repo.PutUser(domain.WorkUserState{GuildID: "guild-1", UserID: "user-1", State: domain.WorkIdleState, Energy: 1, Initialized: true})
	module := NewModuleWithStartRepository(repo, nil, fixedClock{now: time.Unix(100, 0)})
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID(workStartCustomID(item, "user-1"))

	if err := module.StartHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("start handler: %v", err)
	}
	if len(responder.Updates) != 1 || !strings.Contains(responder.Updates[0].Embeds[0].Title, "你的精力不夠") {
		t.Fatalf("unexpected energy error: %#v", responder.Updates)
	}
}

func TestWorkStartBusyPromptsLegacyOverride(t *testing.T) {
	repo, item := workTestRepository()
	repo.PutUser(domain.WorkUserState{GuildID: "guild-1", UserID: "user-1", State: "舊工作", Energy: 9, Initialized: true})
	module := NewModuleWithStartRepository(repo, nil, fixedClock{now: time.Unix(100, 0)})
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID(workStartCustomID(item, "user-1"))

	if err := module.StartHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("start handler: %v", err)
	}
	if len(responder.Updates) != 1 || !strings.Contains(responder.Updates[0].Embeds[0].Title, "你目前有其他工作") {
		t.Fatalf("unexpected busy warning: %#v", responder.Updates)
	}
	buttons := responder.Updates[0].Components[0].Components
	if len(buttons) != 2 || buttons[0].Label != "是" || buttons[0].Emoji != "✅" || !strings.HasPrefix(buttons[0].CustomID, "mhcat:v1:work:override:") {
		t.Fatalf("unexpected override button: %#v", buttons)
	}
	if buttons[1].Label != "否" || buttons[1].Emoji != "❎" || !strings.HasPrefix(buttons[1].CustomID, "mhcat:v1:work:cancel:") {
		t.Fatalf("unexpected cancel button: %#v", buttons)
	}
}

func TestWorkOverrideSuccess(t *testing.T) {
	repo, item := workTestRepository()
	repo.PutUser(domain.WorkUserState{GuildID: "guild-1", UserID: "user-1", State: "舊工作", Energy: 9, Initialized: true})
	module := NewModuleWithStartRepository(repo, nil, fixedClock{now: time.Unix(100, 0)})
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID(workOverrideCustomID(item, "user-1"))

	if err := module.OverrideHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("override handler: %v", err)
	}
	if len(responder.Updates) != 1 || responder.Updates[0].Embeds[0].Title != "<:working:1048617967799242772> 成功取得該工作!" {
		t.Fatalf("unexpected override update: %#v", responder.Updates)
	}
	user, err := repo.GetWorkUser(context.Background(), "guild-1", "user-1")
	if err != nil {
		t.Fatalf("get updated user: %v", err)
	}
	if user.State != "礦坑" || user.Energy != 6 {
		t.Fatalf("unexpected user after override: %#v", user)
	}
}

func TestWorkCancelUsesLegacyCancelMessage(t *testing.T) {
	_, item := workTestRepository()
	module := NewModule(nil)
	responder := fakediscord.NewResponder()
	interaction := fakediscord.ComponentInteractionFromID(workCancelCustomID(item, "user-1"))

	if err := module.CancelHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("cancel handler: %v", err)
	}
	if len(responder.Updates) != 1 || responder.Updates[0].Content != ":x: | **成功取消!**" {
		t.Fatalf("unexpected cancel update: %#v", responder.Updates)
	}
}

func TestWorkModuleRegistersRoute(t *testing.T) {
	router := interactions.NewRouter()
	module := NewModule(nil)
	if err := module.RegisterRoutes(router); err != nil {
		t.Fatalf("register work route: %v", err)
	}
	responder := fakediscord.NewResponder()
	interaction := fakediscord.SlashInteractionWithOptions(CommandName, subcommandAddWork, nil)
	if err := router.Handle(context.Background(), interaction, responder); err != nil {
		t.Fatalf("route work command: %v", err)
	}
	if len(responder.Replies) != 1 {
		t.Fatalf("expected routed reply, got %#v", responder.Replies)
	}
}

func workTestRepository() (*fakemongo.WorkInterfaceRepository, domain.WorkItem) {
	repo := fakemongo.NewWorkInterfaceRepository()
	repo.PutConfig(domain.WorkConfig{GuildID: "guild-1", MaxEnergy: 20})
	item := domain.WorkItem{GuildID: "guild-1", Name: "礦坑", DurationSec: 3600, EnergyCost: 3, CoinReward: 88}
	repo.PutItems("guild-1", item)
	repo.PutUser(domain.WorkUserState{GuildID: "guild-1", UserID: "user-1", State: domain.WorkIdleState, Energy: 9, Initialized: true})
	return repo, item
}

func workAdminSlash(subcommand string, options map[string]string) interactions.Interaction {
	interaction := fakediscord.SlashInteractionWithOptions(CommandName, subcommand, options)
	interaction.Actor.PermissionBits = permissionManageMessages
	return interaction
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}
