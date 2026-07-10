package poll

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

func pollOutboundMessage(poll domain.Poll, memberCount int, color int) ports.OutboundMessage {
	return pollOutboundMessageWithChangeText(poll, memberCount, color, changeText(poll))
}

func initialPollOutboundMessage(poll domain.Poll, memberCount int, color int) ports.OutboundMessage {
	message := pollOutboundMessageWithChangeText(poll, memberCount, color, "不能")
	message.Components = initialPollOutboundComponents(poll)
	return message
}

func pollOutboundMessageWithChangeText(poll domain.Poll, memberCount int, color int, change string) ports.OutboundMessage {
	return ports.OutboundMessage{
		Embeds: []ports.OutboundEmbed{{
			Title:       pollEmbedTitle(poll),
			Description: pollEmbedDescription(poll, memberCount, change),
			Color:       color,
		}},
		Components: pollOutboundComponents(poll),
	}
}

func pollEmbedTitle(poll domain.Poll) string {
	return "<:poll:1023968837965709312> | 投票\n" + poll.Question
}

func pollEmbedDescription(poll domain.Poll, memberCount int, change string) string {
	voters := poll.UniqueVoterCount()
	percentage := "0.00"
	if memberCount > 0 {
		percentage = fmt.Sprintf("%.2f", float64(voters)/float64(memberCount)*100)
	}
	return fmt.Sprintf(`<:vote:1023969411369025576> **總投票人數:`+"`%d` / `%d`|參與率:`%s`%%**\n\n"+`<:YellowSmallDot:1023970607429328946> **每人可以投給`+"`%d`"+`個選項
<:YellowSmallDot:1023970607429328946> `+"`%s`"+`改投其他選項
<:YellowSmallDot:1023970607429328946> `+"`%s`"+`看到投票結果
<:YellowSmallDot:1023970607429328946> `+"`%s`"+`投票**
`, voters, memberCount, percentage, poll.MaxChoices, change, resultText(poll), anonymousText(poll))
}

func changeText(poll domain.Poll) string {
	if poll.CanChangeChoice {
		return "可以"
	}
	return "無法"
}

func resultText(poll domain.Poll) string {
	if poll.CanSeeResult {
		return "可以"
	}
	return "無法"
}

func anonymousText(poll domain.Poll) string {
	if poll.Anonymous {
		return "匿名"
	}
	return "實名"
}

func pollOutboundComponents(poll domain.Poll) []ports.OutboundComponentRow {
	return pollOutboundComponentsWithOwnerMenu(poll, ownerMenuComponent(poll))
}

func initialPollOutboundComponents(poll domain.Poll) []ports.OutboundComponentRow {
	return pollOutboundComponentsWithOwnerMenu(poll, initialOwnerMenuComponent(poll))
}

func pollOutboundComponentsWithOwnerMenu(poll domain.Poll, ownerMenu ports.OutboundComponent) []ports.OutboundComponentRow {
	rows := make([]ports.OutboundComponentRow, 0, 5)
	current := ports.OutboundComponentRow{}
	for index, choice := range poll.Choices {
		if len(current.Components) == 5 {
			rows = append(rows, current)
			current = ports.OutboundComponentRow{}
		}
		current.Components = append(current.Components, ports.OutboundComponent{
			Type:     "button",
			CustomID: versionedPollID("vote", map[string]string{"i": strconv.Itoa(index)}),
			Label:    choice,
			Style:    "secondary",
		})
	}
	resultButton := ports.OutboundComponent{
		Type:     "button",
		CustomID: versionedPollID("result", nil),
		Label:    "查看投票結果",
		Emoji:    "<:analysis:1023965999357243432>",
		Style:    "success",
	}
	if len(current.Components) == 5 {
		rows = append(rows, current)
		current = ports.OutboundComponentRow{}
	}
	current.Components = append(current.Components, resultButton)
	rows = append(rows, current)
	rows = append(rows, ports.OutboundComponentRow{Components: []ports.OutboundComponent{ownerMenu}})
	return rows
}

func initialOwnerMenuComponent(poll domain.Poll) ports.OutboundComponent {
	component := ownerMenuComponent(poll)
	component.Options[2].Label = "允許變更選項"
	component.Options[4].Label = "結束投票"
	component.Options[4].Description = "讓該投票變為無法再變更選項或投票(可再次開啟)"
	return component
}

func ownerMenuComponent(poll domain.Poll) ports.OutboundComponent {
	return ports.OutboundComponent{
		Type:        "select",
		CustomID:    versionedPollID("owner_menu", nil),
		Placeholder: "🔧投票發起人操作",
		Options: []ports.OutboundSelectOption{
			{
				Label:       publicResultLabel(poll),
				Description: publicResultDescription(poll),
				Value:       string(domain.PollTogglePublicResult),
				Emoji:       "<:publicrelation:1023972880385585212>",
			},
			{
				Label:       "啟用多選投票",
				Description: "讓所有成員都可以投票超過1個以上",
				Value:       "poll_can_choose_many",
				Emoji:       "<:maybe:1023971826948391074>",
			},
			{
				Label:       changeChoiceLabel(poll),
				Description: changeChoiceDescription(poll),
				Value:       string(domain.PollToggleChangeChoice),
				Emoji:       "<:exchange:1023972882046525491>",
			},
			{
				Label:       "改為匿名投票",
				Description: "讓所有無法得知有誰參加抽獎",
				Value:       string(domain.PollToggleAnonymous),
				Emoji:       "<:unknown:1024241985583853598>",
			},
			{
				Label:       endPollLabel(poll),
				Description: endPollDescription(poll),
				Value:       string(domain.PollToggleEnd),
				Emoji:       endPollEmoji(poll),
			},
			{
				Label:       "匯出為excel檔",
				Description: "如果成員過多的話可以使用這個查看誰投票",
				Value:       "poll_excel_result",
				Emoji:       "<:sheets:1023972957330100324>",
			},
		},
	}
}

func publicResultLabel(poll domain.Poll) string {
	if poll.CanSeeResult {
		return "隱藏投票結果"
	}
	return "公開投票結果"
}

func publicResultDescription(poll domain.Poll) string {
	if poll.CanSeeResult {
		return "讓所有成員都無法查看該投票結果"
	}
	return "讓所有成員都可以查看該投票結果"
}

func changeChoiceLabel(poll domain.Poll) string {
	if poll.CanChangeChoice {
		return "無法變更選項"
	}
	return "可以變更選項"
}

func changeChoiceDescription(poll domain.Poll) string {
	if poll.CanChangeChoice {
		return "讓所有成員都無法更改投票選項"
	}
	return "讓所有成員都可以更改投票選項"
}

func endPollLabel(poll domain.Poll) string {
	if poll.Ended {
		return "重啟投票"
	}
	return "終止投票"
}

func endPollDescription(poll domain.Poll) string {
	if poll.Ended {
		return "讓投票可以繼續使用讓"
	}
	return "該投票變為無法再變更選項或投票(可再次開啟)讓"
}

func endPollEmoji(poll domain.Poll) string {
	if poll.Ended {
		return "<:playbutton:1023972876921081947>"
	}
	return "<:stop:1023972878678503434>"
}

func maxChoiceMenuMessage(messageID string, choiceCount int, color int) responses.Message {
	options := make([]responses.SelectOption, 0, choiceCount-1)
	for index := 1; index < choiceCount; index++ {
		options = append(options, responses.SelectOption{
			Label:       fmt.Sprintf("%d個選項", index),
			Description: fmt.Sprintf("最多可以可以投給%d個選項", index),
			Value:       strconv.Itoa(index),
		})
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<:maybe:1023971826948391074> | 請選擇最多選擇數量",
			Color: color,
		}},
		Components: []responses.ComponentRow{{Components: []responses.Component{{
			Type:        responses.ComponentTypeSelect,
			CustomID:    versionedPollID("max_choices", map[string]string{"m": messageID}),
			Placeholder: "請選擇可以最多選擇數!",
			Options:     options,
		}}}},
	}
}

func pollResultEmbedMessage(poll domain.Poll, color int) responses.Message {
	fields := make([]responses.EmbedField, 0, len(poll.Choices))
	for _, choice := range poll.Choices {
		count := poll.CountChoice(choice)
		percentage := "0.00"
		if len(poll.Votes) > 0 {
			percentage = fmt.Sprintf("%.2f", float64(count)/float64(len(poll.Votes))*100)
		}
		fields = append(fields, responses.EmbedField{
			Name:   fmt.Sprintf("%s(共%d人 `%s`%%)", choice, count, percentage),
			Value:  pollChoiceVoters(poll, choice),
			Inline: false,
		})
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:  "<:poll:1023968837965709312> | " + poll.Question,
			Fields: fields,
			Color:  color,
		}},
	}
}

func pollChoiceVoters(poll domain.Poll, choice string) string {
	if poll.Anonymous && len(poll.Votes) > 0 {
		return "該投票為匿名，無法查看誰有進行投票"
	}
	if len(poll.Votes) > 50 {
		return "由於人數過多，無法顯示所有人"
	}
	var voters []string
	for _, vote := range poll.Votes {
		if vote.Choice == choice {
			voters = append(voters, "<@"+vote.UserID+">")
		}
	}
	if len(voters) == 0 {
		return "<a:Discord_AnimatedNo:1015989839809757295> | 還沒有人投給這個選項"
	}
	return strings.Join(voters, "")
}

func versionedPollID(action string, values map[string]string) string {
	payload, err := customid.KeyValuePayload(values)
	if err != nil {
		return ""
	}
	id, err := customid.Encode(customid.InteractionKindComponent, "poll", action, payload)
	if err != nil {
		return ""
	}
	return id
}
