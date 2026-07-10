package lottery

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	permissionManageMessages     int64 = 8192
	legacyLotteryGreenColor            = 0x57F287
	legacyLotteryResultColor           = 0x5865F2
	legacyLotteryWinnerLimit           = 50
	legacyLotteryParticipantFile       = "discord.txt"
)

var legacyLotteryBaseIDPattern = regexp.MustCompile(`^[0-9]{13,20}lotter$`)

func (m Module) EnterHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		id, ok := lotteryIDFromAction(interaction.CustomID, "enter")
		if !ok {
			return responder.EditOriginal(ctx, lotteryComponentErrorMessage("很抱歉，出現了錯誤!"))
		}
		_, err := m.service.Join(ctx, domain.LotteryJoinRequest{
			GuildID: interaction.Actor.GuildID,
			ID:      id,
			UserID:  interaction.Actor.UserID,
		}, interaction.Actor.RoleIDs, m.now())
		if err != nil {
			return responder.EditOriginal(ctx, lotteryEnterErrorMessage(err))
		}
		return responder.EditOriginal(ctx, lotterySuccessEmbedMessage("<a:green_tick:994529015652163614> | 成功參加抽獎!"))
	}
}

func (m Module) SearchHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		id, ok := lotteryIDFromAction(interaction.CustomID, "search")
		if !ok {
			return responder.EditOriginal(ctx, lotteryComponentErrorMessage("很抱歉，出現了錯誤!"))
		}
		lottery, err := m.service.Get(ctx, interaction.Actor.GuildID, id)
		if err != nil {
			return responder.EditOriginal(ctx, lotteryLookupErrorMessage(err, false))
		}
		guildOwnerID := m.guildOwnerID(ctx, interaction.Actor.GuildID)
		canManage := m.service.CanManage(lottery, interaction.Actor.UserID, guildOwnerID, interaction.Actor.HasPermission(permissionManageMessages))
		return responder.EditOriginal(ctx, m.lotterySearchMessage(ctx, lottery, interaction.Actor.UserID, canManage))
	}
}

func (m Module) RerollHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		id, ok := lotteryIDFromAction(interaction.CustomID, "reroll")
		if !ok {
			return responder.EditOriginal(ctx, lotteryComponentErrorMessage("很抱歉，出現了錯誤!"))
		}
		guildOwnerID := m.guildOwnerID(ctx, interaction.Actor.GuildID)
		lottery, err := m.service.GetManaged(ctx, interaction.Actor.GuildID, id, interaction.Actor.UserID, guildOwnerID, interaction.Actor.HasPermission(permissionManageMessages))
		if err != nil {
			return responder.EditOriginal(ctx, lotteryManageErrorMessage(err, false))
		}
		message, err := m.lotteryWinnerMessage(lottery)
		if err != nil || m.messages == nil || lottery.ChannelID == "" {
			return responder.EditOriginal(ctx, lotteryComponentErrorMessage("很抱歉，出現了錯誤!"))
		}
		if _, err := m.messages.SendMessage(ctx, lottery.ChannelID, message); err != nil {
			return responder.EditOriginal(ctx, lotteryComponentErrorMessage("很抱歉，出現了錯誤!"))
		}
		if _, err := m.service.EndManaged(ctx, interaction.Actor.GuildID, id, interaction.Actor.UserID, guildOwnerID, interaction.Actor.HasPermission(permissionManageMessages)); err != nil {
			return responder.EditOriginal(ctx, lotteryManageErrorMessage(err, false))
		}
		return responder.EditOriginal(ctx, lotteryComponentMessage("<a:green_tick:994529015652163614> | 成功重抽!"))
	}
}

func (m Module) StopHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		id, ok := lotteryIDFromAction(interaction.CustomID, "stop")
		if !ok {
			return responder.EditOriginal(ctx, lotteryComponentErrorMessage("很抱歉，出現了錯誤!"))
		}
		guildOwnerID := m.guildOwnerID(ctx, interaction.Actor.GuildID)
		if _, err := m.service.EndManaged(ctx, interaction.Actor.GuildID, id, interaction.Actor.UserID, guildOwnerID, interaction.Actor.HasPermission(permissionManageMessages)); err != nil {
			return responder.EditOriginal(ctx, lotteryManageErrorMessage(err, true))
		}
		return responder.EditOriginal(ctx, lotterySuccessEmbedMessage("<a:green_tick:994529015652163614> | 成功取消此次抽獎!"))
	}
}

func (m Module) guildOwnerID(ctx context.Context, guildID string) string {
	if m.discord == nil || strings.TrimSpace(guildID) == "" {
		return ""
	}
	guild, err := m.discord.GuildInfo(ctx, guildID)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(guild.OwnerID)
}

func (m Module) lotterySearchMessage(ctx context.Context, lottery domain.Lottery, actorUserID string, canManage bool) responses.Message {
	ids := make([]string, 0, len(lottery.Participants))
	for _, participant := range lottery.Participants {
		ids = append(ids, participant.UserID)
	}
	tags := map[string]string{}
	if m.members != nil {
		if resolved, err := m.members.MemberTags(ctx, lottery.GuildID, ids); err == nil {
			tags = resolved
		}
	}
	displayNames := make([]string, 0, len(lottery.Participants))
	fileLines := make([]string, 0, len(lottery.Participants))
	joined := false
	for _, participant := range lottery.Participants {
		tag, found := tags[participant.UserID]
		if !found || strings.TrimSpace(tag) == "" {
			displayNames = append(displayNames, "使用者已消失!")
			tag = "使用者已退出伺服器!"
		} else {
			tag = legacyLotteryParticipantTag(tag)
			displayNames = append(displayNames, tag)
		}
		fileLines = append(fileLines, tag+"(id:"+participant.UserID+")|參加時間:"+legacyLotteryParticipantTime(participant))
		if participant.UserID == strings.TrimSpace(actorUserID) {
			joined = true
		}
	}
	participantText := "┃ " + strings.Join(displayNames, " ┃ ") + "┃"
	if len(displayNames) >= 100 {
		participantText = "**由於人數過多，無法顯示所有成員名稱!\n請使用`.txt`檔案觀看**"
	}
	joinedText := "`沒有`"
	if joined {
		joinedText = "`有`"
	}
	message := responses.Message{
		Embeds: []responses.Embed{{
			Title: "抽獎人數資訊",
			Description: "<:list:992002476360343602>**目前共有**`" + strconv.Itoa(len(lottery.Participants)) + "`**人參加抽獎**\n" +
				"<:star:987020551698649138>**您是否有參加該抽獎:**" + joinedText + "\n\n" + participantText,
			Color: m.color(),
		}},
		Files: []responses.File{{
			Name:        legacyLotteryParticipantFile,
			ContentType: "text/plain; charset=utf-8",
			Data:        []byte(strings.Join(fileLines, "\n")),
		}},
		AllowedMentions: &responses.AllowedMentions{},
		Ephemeral:       true,
	}
	if canManage {
		message.Content = "<:shield:1019529265101930567> | 你有權限(創辦人或服主)執行終止抽獎或是重抽，是否要執行其中一項?"
		message.Components = []responses.ComponentRow{{Components: []responses.Component{
			{Type: responses.ComponentTypeButton, CustomID: lottery.ID + "restart", Label: "點我重抽!", Emoji: "<:votingbox:988878045882499092>", Style: responses.ButtonStyleSuccess},
			{Type: responses.ComponentTypeButton, CustomID: lottery.ID + "stop", Label: "點我取消此次抽獎!", Emoji: "<:warning:985590881698590730>", Style: responses.ButtonStyleDanger},
		}}}
	}
	return message
}

func (m Module) lotteryWinnerMessage(lottery domain.Lottery) (ports.OutboundMessage, error) {
	winnerCount := lottery.WinnerCount
	if winnerCount < 0 {
		winnerCount = 0
	}
	if winnerCount > legacyLotteryWinnerLimit {
		winnerCount = legacyLotteryWinnerLimit
	}
	winners := make([]string, 0, winnerCount)
	if len(lottery.Participants) > 0 {
		for range winnerCount {
			index, err := m.randomIndex(len(lottery.Participants))
			if err != nil || index < 0 || index >= len(lottery.Participants) {
				return ports.OutboundMessage{}, errors.New("draw lottery winner")
			}
			winners = append(winners, lottery.Participants[index].UserID)
		}
	}
	description := "**沒有人參加抽獎欸QQ**"
	content := ""
	if len(winners) > 0 {
		content = "<@" + strings.Join(winners, "><@") + ">"
		description = "\n**<:celebration:997374188060946495> 恭喜:**\n<@" + strings.Join(winners, ">\n<@") + ">\n<:gift:994585975445528576> **抽中:** " + lottery.Gift + "\n"
	}
	return ports.OutboundMessage{
		Content: content,
		Embeds: []ports.OutboundEmbed{{
			Title:       "<:fireworks:997374182016958494> 恭喜中獎者! <:fireworks:997374182016958494>",
			Description: description,
			Color:       legacyLotteryResultColor,
			FooterText:  "沒抽中的我給你一個擁抱w",
		}},
		AllowedMentions: ports.AllowedMentions{UserIDs: uniqueLotteryUserIDs(winners)},
	}, nil
}

func lotteryIDFromAction(raw string, action string) (string, bool) {
	raw = strings.TrimSpace(raw)
	suffix := ""
	switch action {
	case "enter":
	case "search":
		suffix = "search"
	case "reroll":
		suffix = "restart"
	case "stop":
		suffix = "stop"
	default:
		return "", false
	}
	if suffix != "" {
		if !strings.HasSuffix(raw, suffix) {
			return "", false
		}
		raw = strings.TrimSuffix(raw, suffix)
	}
	return raw, legacyLotteryBaseIDPattern.MatchString(raw)
}

func legacyLotteryParticipantTime(participant domain.LotteryParticipant) string {
	if participant.JoinedAtMillis > 0 {
		location, err := time.LoadLocation("Asia/Taipei")
		if err == nil {
			return time.UnixMilli(participant.JoinedAtMillis).In(location).Format("2006/01/02\u200915:04:05") + " [台北標準時間]"
		}
	}
	return participant.JoinedAtRaw
}

func legacyLotteryParticipantTag(tag string) string {
	if strings.Contains(tag, "#") {
		return tag
	}
	return tag + "#0"
}

func uniqueLotteryUserIDs(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func lotteryEnterErrorMessage(err error) responses.Message {
	switch {
	case errors.Is(err, ports.ErrLotteryNotFound):
		return lotteryComponentMessage("<a:Discord_AnimatedNo:1015989839809757295> | 這個抽獎已經因為時間過久而被刪除資料(結束超過30天)!")
	case errors.Is(err, ports.ErrLotteryEnded):
		return lotteryComponentMessage("<a:Discord_AnimatedNo:1015989839809757295> | 很抱歉，這個抽獎已經過期!")
	case errors.Is(err, ports.ErrLotteryAlreadyJoined):
		return lotteryComponentMessage("<a:Discord_AnimatedNo:1015989839809757295> | 你無法重複參加!")
	case errors.Is(err, ports.ErrLotteryFull):
		return lotteryComponentMessage("<a:Discord_AnimatedNo:1015989839809757295> | 以達到最高參與人數")
	case errors.Is(err, ports.ErrLotteryRoleDenied):
		return lotteryComponentMessage("<a:Discord_AnimatedNo:1015989839809757295> | 很抱歉，創辦人設定你不能抽獎!")
	default:
		return lotteryComponentErrorMessage("很抱歉，出現了錯誤!")
	}
}

func lotteryLookupErrorMessage(err error, stop bool) responses.Message {
	if errors.Is(err, ports.ErrLotteryNotFound) {
		if stop {
			return lotteryComponentMessage("<a:Discord_AnimatedNo:1015989839809757295> | 很抱歉，這個抽獎已經因為超過時間而刪除資料了!")
		}
		return lotteryComponentMessage("<a:Discord_AnimatedNo:1015989839809757295> | 這個抽獎已經因為時間過久而被刪除資料(結束超過30天)!")
	}
	return lotteryComponentErrorMessage("很抱歉，出現了錯誤!")
}

func lotteryManageErrorMessage(err error, stop bool) responses.Message {
	if errors.Is(err, ports.ErrLotteryManagerOnly) {
		return lotteryComponentMessage("<a:Discord_AnimatedNo:1015989839809757295> | 你沒有權限執行這個操作!")
	}
	return lotteryLookupErrorMessage(err, stop)
}

func lotteryComponentMessage(content string) responses.Message {
	return responses.Message{Content: content, AllowedMentions: &responses.AllowedMentions{}, Ephemeral: true}
}

func lotteryComponentErrorMessage(content string) responses.Message {
	return lotteryComponentMessage("<a:Discord_AnimatedNo:1015989839809757295> | " + content)
}

func lotterySuccessEmbedMessage(title string) responses.Message {
	return responses.Message{
		Embeds:          []responses.Embed{{Title: title, Color: legacyLotteryGreenColor}},
		AllowedMentions: &responses.AllowedMentions{},
		Ephemeral:       true,
	}
}
