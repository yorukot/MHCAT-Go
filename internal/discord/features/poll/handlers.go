package poll

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

var (
	ErrPollRepositoryNotConfigured  = errors.New("poll repository is not configured")
	ErrPollSideEffectsNotConfigured = errors.New("poll side-effect ports are not configured")
)

const (
	permissionManageMessages = 8192
	pollSuccessColor         = 0x57F287
	pollErrorColor           = 0xED4245
)

func (m Module) CreateHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.repo == nil {
			return ErrPollRepositoryNotConfigured
		}
		if m.messages == nil {
			return ErrPollSideEffectsNotConfigured
		}
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(m.memberPerm) {
			return responder.EditOriginal(ctx, pollErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		question := interaction.Options["問題"]
		choicesRaw := interaction.Options["選項"]
		choices, msg := validatePollInput(question, choicesRaw)
		if msg != "" {
			return responder.EditOriginal(ctx, pollErrorMessage(msg))
		}
		memberCount := m.memberCount(ctx, interaction.Actor.GuildID, 0)
		color := m.nextPollColor()
		draft := domain.NewPoll(domain.PollCreate{
			GuildID:   interaction.Actor.GuildID,
			MessageID: "pending",
			Question:  question,
			CreatorID: interaction.Actor.UserID,
			Choices:   choices,
		})
		sent, err := m.messages.SendMessage(ctx, interaction.ChannelID, initialPollOutboundMessage(draft, memberCount, color))
		if err != nil {
			return err
		}
		created, err := m.repo.CreatePoll(ctx, domain.PollCreate{
			GuildID:   interaction.Actor.GuildID,
			MessageID: sent.MessageID,
			Question:  question,
			CreatorID: interaction.Actor.UserID,
			Choices:   choices,
		})
		if err != nil {
			_ = m.messages.DeleteMessage(ctx, sent)
			return err
		}
		if sent.MessageID != "" {
			_ = m.messages.EditMessage(ctx, sent, initialPollOutboundMessage(created, memberCount, color))
		}
		if err := responder.EditOriginal(ctx, responses.Message{
			Embeds: []responses.Embed{{
				Title: "<a:green_tick:994529015652163614> | 成功創建投票!",
				Color: pollSuccessColor,
			}},
			Ephemeral: true,
		}); err != nil {
			return err
		}
		return nil
	}
}

func (m Module) VoteHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.repo == nil {
			return ErrPollRepositoryNotConfigured
		}
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		messageID := interaction.MessageID
		poll, err := m.repo.GetPoll(ctx, interaction.Actor.GuildID, messageID)
		if err != nil {
			return responder.EditOriginal(ctx, pollErrorFor(err))
		}
		choice, err := choiceFromInteraction(interaction, poll)
		if err != nil {
			return responder.EditOriginal(ctx, pollErrorMessage("很抱歉，找不到這個投票選項!"))
		}
		change, err := m.repo.Vote(ctx, interaction.Actor.GuildID, messageID, interaction.Actor.UserID, choice, domain.LegacyVoteTime(m.now()))
		if err != nil {
			return responder.EditOriginal(ctx, pollErrorFor(err))
		}
		m.refreshPollMessage(ctx, interaction, interaction.MessageID, change.Poll)
		title := "<a:green_tick:994529015652163614> | 你成功投給`" + choice + "`!"
		description := "如需更改選項，可以再點選一次該選項即可取消投票"
		if change.Removed {
			title = "<a:green_tick:994529015652163614> | 成功取消投給`" + choice + "`!"
			description = ""
		}
		if err := responder.EditOriginal(ctx, responses.Message{
			Embeds: []responses.Embed{{Title: title, Description: description, Color: pollSuccessColor}},
		}); err != nil {
			return err
		}
		return nil
	}
}

func (m Module) ResultHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.repo == nil {
			return ErrPollRepositoryNotConfigured
		}
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		poll, err := m.repo.GetPoll(ctx, interaction.Actor.GuildID, interaction.MessageID)
		if err != nil {
			return responder.EditOriginal(ctx, pollErrorFor(err))
		}
		if !poll.CanSeeResult && poll.CreatorID != interaction.Actor.UserID {
			return responder.EditOriginal(ctx, pollErrorMessageWithReason("這個投票不是公開的!", "如需公開該投票，請使用下方選擇器!"))
		}
		if len(poll.Votes) == 0 {
			return responder.EditOriginal(ctx, pollErrorMessage("還沒有人參與投票!"))
		}
		if err := responder.EditOriginal(ctx, pollResultMessage(ctx, poll, m.members, m.nextPollColor())); err != nil {
			return err
		}
		m.refreshPollMessage(ctx, interaction, interaction.MessageID, poll)
		return nil
	}
}

func (m Module) OwnerMenuHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.repo == nil {
			return ErrPollRepositoryNotConfigured
		}
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		poll, err := m.repo.GetPoll(ctx, interaction.Actor.GuildID, interaction.MessageID)
		if err != nil {
			return responder.EditOriginal(ctx, pollErrorFor(err))
		}
		if poll.CreatorID != interaction.Actor.UserID {
			return responder.EditOriginal(ctx, pollErrorMessage("你不是投票發起人，無法操作!"))
		}
		value := firstValue(interaction)
		if value == "poll_can_choose_many" {
			if len(poll.Choices) < 3 {
				return responder.EditOriginal(ctx, pollErrorMessage("必須要有3個選項才能啟用多選"))
			}
			if err := responder.EditOriginal(ctx, maxChoiceMenuMessage(interaction.MessageID, len(poll.Choices), m.nextPollColor())); err != nil {
				return err
			}
			m.refreshPollMessage(ctx, interaction, interaction.MessageID, poll)
			return nil
		}
		if value == "poll_excel_result" {
			if poll.Anonymous {
				return responder.EditOriginal(ctx, pollErrorMessage("該投票為匿名，無法查看投票資訊!"))
			}
			message, err := pollExcelMessage(ctx, poll, m.members)
			if err != nil {
				return err
			}
			if err := responder.EditOriginal(ctx, message); err != nil {
				return err
			}
			m.refreshPollMessage(ctx, interaction, interaction.MessageID, poll)
			return nil
		}
		toggle := domain.PollToggle(value)
		updated, err := m.repo.TogglePoll(ctx, interaction.Actor.GuildID, interaction.MessageID, toggle)
		if err != nil {
			if errors.Is(err, ports.ErrPollAnonymousLocked) {
				if responseErr := responder.EditOriginal(ctx, pollErrorFor(err)); responseErr != nil {
					return responseErr
				}
				m.refreshPollMessage(ctx, interaction, interaction.MessageID, poll)
				return nil
			}
			return responder.EditOriginal(ctx, pollErrorFor(err))
		}
		m.refreshPollMessage(ctx, interaction, interaction.MessageID, updated)
		return responder.EditOriginal(ctx, pollDoneMessage(doneMessageForToggle(toggle, updated)))
	}
}

func (m Module) MaxChoicesHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.repo == nil {
			return ErrPollRepositoryNotConfigured
		}
		messageID, err := maxChoiceMessageID(interaction)
		if err != nil || messageID == "" {
			return responder.Reply(ctx, pollErrorMessage("這個多選選單已經無法安全對應原始投票，請重新從投票發起人選單操作。"))
		}
		maxChoices, err := strconv.Atoi(firstValue(interaction))
		if err != nil || maxChoices < 1 {
			return responder.UpdateMessage(ctx, pollErrorMessage("請選擇正確的最多選擇數量"))
		}
		updated, err := m.repo.SetMaxChoices(ctx, interaction.Actor.GuildID, messageID, maxChoices)
		if err != nil {
			return responder.UpdateMessage(ctx, pollErrorFor(err))
		}
		m.refreshPollMessage(ctx, interaction, messageID, updated)
		return responder.UpdateMessage(ctx, responses.Message{
			Embeds: []responses.Embed{{
				Title: fmt.Sprintf("<a:green_tick:994529015652163614> | 成功將最多選擇數量設為%d", maxChoices),
				Color: m.nextPollColor(),
			}},
		})
	}
}

func validatePollInput(question string, choicesRaw string) ([]string, string) {
	if len(utf16.Encode([]rune(question))) > 2500 {
		return nil, "問題字數不可超過2500"
	}
	parts := strings.Split(choicesRaw, "^")
	if len(parts) < 2 {
		return nil, "最少需要2個選項!"
	}
	if len(parts) > 19 {
		return nil, "最多只能有19個選項!"
	}
	seen := map[string]struct{}{}
	for _, part := range parts {
		if _, ok := seen[part]; ok {
			return nil, "選項名稱不可以重複!"
		}
		seen[part] = struct{}{}
	}
	choices := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(utf16.Encode([]rune(part))) > 80 {
			return nil, "你輸入的選項字數不能超過80"
		}
		if part == "" {
			return nil, "^跟^中間請填入選項，不可為空"
		}
		choices = append(choices, part)
	}
	return choices, ""
}

func choiceFromInteraction(interaction interactions.Interaction, poll domain.Poll) (string, error) {
	if strings.HasPrefix(interaction.CustomID, "poll_") {
		choice := strings.TrimPrefix(interaction.CustomID, "poll_")
		if choice == "" {
			return "", ports.ErrPollChoiceNotFound
		}
		return choice, nil
	}
	parsed, err := customid.ParseComponent(interaction.CustomID)
	if err != nil {
		return "", err
	}
	indexValue := parsed.Payload.Values["i"]
	index, err := strconv.Atoi(indexValue)
	if err != nil || index < 0 || index >= len(poll.Choices) {
		return "", ports.ErrPollChoiceNotFound
	}
	return poll.Choices[index], nil
}

func maxChoiceMessageID(interaction interactions.Interaction) (string, error) {
	parsed, err := customid.ParseComponent(interaction.CustomID)
	if err != nil {
		return "", err
	}
	return parsed.Payload.Values["m"], nil
}

func firstValue(interaction interactions.Interaction) string {
	if len(interaction.Values) == 0 {
		return ""
	}
	return strings.TrimSpace(interaction.Values[0])
}

func (m Module) memberCount(ctx context.Context, guildID string, fallback int) int {
	if m.members == nil {
		return fallback
	}
	count, err := m.members.CountNonBotMembers(ctx, guildID)
	if err != nil || count < 0 {
		return fallback
	}
	if count == 0 && fallback > 0 {
		return fallback
	}
	return count
}

func (m Module) refreshPollMessage(ctx context.Context, interaction interactions.Interaction, messageID string, poll domain.Poll) {
	if m.messages == nil || interaction.ChannelID == "" || messageID == "" {
		return
	}
	count := m.memberCount(ctx, interaction.Actor.GuildID, poll.UniqueVoterCount())
	_ = m.messages.EditMessage(ctx, ports.MessageRef{ChannelID: interaction.ChannelID, MessageID: messageID}, pollOutboundMessage(poll, count, m.nextPollColor()))
}

func (m Module) now() time.Time {
	if m.clock == nil {
		return time.Now()
	}
	return m.clock.Now()
}

func pollErrorFor(err error) responses.Message {
	switch {
	case errors.Is(err, ports.ErrPollNotFound):
		return pollErrorMessage("該投票已經過期(超過30天會自動刪除)")
	case errors.Is(err, ports.ErrPollEnded):
		return pollErrorMessage("該投票已被結束!")
	case errors.Is(err, ports.ErrPollChangeNotAllowed):
		return pollErrorMessage("很抱歉，該投票不支援更改選項!")
	case errors.Is(err, ports.ErrPollChoiceLimit):
		return pollErrorMessageWithReason("你已經達到該投票最大上限", "如需更改選項，請將原來所選的選項點掉!")
	case errors.Is(err, ports.ErrPollAnonymousLocked):
		return pollErrorMessage("匿名的投票無法改為實名!")
	default:
		return pollErrorMessage("很抱歉，出現了錯誤!")
	}
}

func pollErrorMessage(content string) responses.Message {
	return pollErrorMessageWithReason(content, "")
}

func pollErrorMessageWithReason(content string, reason string) responses.Message {
	return responses.Message{
		Ephemeral: true,
		Embeds: []responses.Embed{{
			Title:       "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Description: reason,
			Color:       pollErrorColor,
		}},
	}
}

func pollDoneMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:green_tick:994529015652163614>" + content,
			Color: pollSuccessColor,
		}},
	}
}

func doneMessageForToggle(toggle domain.PollToggle, poll domain.Poll) string {
	switch toggle {
	case domain.PollTogglePublicResult:
		if poll.CanSeeResult {
			return "成功將投票結果設為公開!"
		}
		return "成功將投票結果設為不公開!"
	case domain.PollToggleChangeChoice:
		if poll.CanChangeChoice {
			return "成功將投票設為可以更改選項!"
		}
		return "成功將投票設為無法更改選項!"
	case domain.PollToggleAnonymous:
		return "成功將投票設為匿名投票!"
	case domain.PollToggleEnd:
		if poll.Ended {
			return "成功結束投票!"
		}
		return "成功重新開啟投票!"
	default:
		return "成功更新投票!"
	}
}
