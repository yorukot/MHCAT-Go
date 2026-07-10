package poll

import (
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

func TestPollDefinitionMatchesLegacyVisibleContract(t *testing.T) {
	permissions := manageMessagesPermission
	want := []commands.Definition{{
		Type:                     commands.CommandTypeChatInput,
		Name:                     "投票創建",
		Description:              "創建一個萬能的投票",
		DefaultMemberPermissions: &permissions,
		Ownership:                commands.ManagedOwnership("poll-wave-a", commands.ScopeGuild),
		Options: []commands.Option{
			{
				Type:        commands.OptionTypeString,
				Name:        "問題",
				Description: "輸入你要問的問題!ex:我要買甚麼?",
				Required:    true,
			},
			{
				Type:        commands.OptionTypeString,
				Name:        "選項",
				Description: "輸入回答的選項，請用^將各個選項分開 ex:電腦^手機^兩個都要^!",
				Required:    true,
			},
		},
	}}
	if got := Definitions(); !reflect.DeepEqual(got, want) {
		t.Fatalf("definitions = %#v, want %#v", got, want)
	}
}

func TestInitialPollMessageMatchesAuditedVisibleContract(t *testing.T) {
	poll := domain.NewPoll(domain.PollCreate{
		GuildID:   "guild-1",
		MessageID: "message-1",
		Question:  "今天吃什麼?",
		CreatorID: "owner-1",
		Choices:   []string{"拉麵", "壽司"},
	})
	want := ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title: "<:poll:1023968837965709312> | 投票\n今天吃什麼?",
			Description: `<:vote:1023969411369025576> **總投票人數:` + "`0` / `10`|參與率:`0.00`%**\n\n" + `<:YellowSmallDot:1023970607429328946> **每人可以投給` + "`1`" + `個選項
<:YellowSmallDot:1023970607429328946> ` + "`不能`" + `改投其他選項
<:YellowSmallDot:1023970607429328946> ` + "`無法`" + `看到投票結果
<:YellowSmallDot:1023970607429328946> ` + "`實名`" + `投票**
`,
			Color: 0x123456,
		}},
		Components: []ports.OutboundComponentRow{
			{Components: []ports.OutboundComponent{
				{Type: "button", CustomID: "mhcat:v1:poll:vote:i=0", Label: "拉麵", Style: "secondary"},
				{Type: "button", CustomID: "mhcat:v1:poll:vote:i=1", Label: "壽司", Style: "secondary"},
				{Type: "button", CustomID: "mhcat:v1:poll:result:", Label: "查看投票結果", Emoji: "<:analysis:1023965999357243432>", Style: "success"},
			}},
			{Components: []ports.OutboundComponent{{
				Type:        "select",
				CustomID:    "mhcat:v1:poll:owner_menu:",
				Placeholder: "🔧投票發起人操作",
				Options: []ports.OutboundSelectOption{
					{Label: "公開投票結果", Description: "讓所有成員都可以查看該投票結果", Value: "poll_public_result", Emoji: "<:publicrelation:1023972880385585212>"},
					{Label: "啟用多選投票", Description: "讓所有成員都可以投票超過1個以上", Value: "poll_can_choose_many", Emoji: "<:maybe:1023971826948391074>"},
					{Label: "允許變更選項", Description: "讓所有成員都可以更改投票選項", Value: "poll_can_change_choose", Emoji: "<:exchange:1023972882046525491>"},
					{Label: "改為匿名投票", Description: "讓所有無法得知有誰參加抽獎", Value: "poll_anonymous", Emoji: "<:unknown:1024241985583853598>"},
					{Label: "結束投票", Description: "讓該投票變為無法再變更選項或投票(可再次開啟)", Value: "poll_end_poll", Emoji: "<:stop:1023972878678503434>"},
					{Label: "匯出為excel檔", Description: "如果成員過多的話可以使用這個查看誰投票", Value: "poll_excel_result", Emoji: "<:sheets:1023972957330100324>"},
				},
			}}},
		},
	}
	if got := initialPollOutboundMessage(poll, 10, 0x123456); !reflect.DeepEqual(got, want) {
		t.Fatalf("initial poll = %#v, want %#v", got, want)
	}
}

func TestPollFixedMessagesMatchAuditedLegacyContract(t *testing.T) {
	tests := []struct {
		name string
		got  responses.Message
		want responses.Message
	}{
		{
			name: "error",
			got:  pollErrorMessage("該投票已被結束!"),
			want: responses.Message{Ephemeral: true, Embeds: []responses.Embed{{
				Title: "<a:Discord_AnimatedNo:1015989839809757295> | 該投票已被結束!",
				Color: 0xED4245,
			}}},
		},
		{
			name: "error reason",
			got:  pollErrorMessageWithReason("你已經達到該投票最大上限", "如需更改選項，請將原來所選的選項點掉!"),
			want: responses.Message{Ephemeral: true, Embeds: []responses.Embed{{
				Title:       "<a:Discord_AnimatedNo:1015989839809757295> | 你已經達到該投票最大上限",
				Description: "如需更改選項，請將原來所選的選項點掉!",
				Color:       0xED4245,
			}}},
		},
		{
			name: "owner success",
			got:  pollDoneMessage("成功結束投票!"),
			want: responses.Message{Embeds: []responses.Embed{{
				Title: "<a:green_tick:994529015652163614>成功結束投票!",
				Color: 0x57F287,
			}}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !reflect.DeepEqual(test.got, test.want) {
				t.Fatalf("message = %#v, want %#v", test.got, test.want)
			}
		})
	}
}
