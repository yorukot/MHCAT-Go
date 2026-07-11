package onboarding

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/onboarding"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const verificationAnswerInputID = "verification_answer"

func (m Module) VerificationHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{Ephemeral: true}); err != nil {
			return err
		}
		result, err := m.flowService.Start(ctx, interaction.Actor.GuildID, interaction.Actor.UserID)
		if err != nil {
			return responder.EditOriginal(ctx, verificationFlowErrorFromError(err))
		}
		msg, err := verificationPromptMessage(result)
		if err != nil {
			return responder.EditOriginal(ctx, verificationFlowErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		return responder.EditOriginal(ctx, msg)
	}
}

func (m Module) VerificationPromptHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if interaction.RouteKey.Legacy {
			answer := strings.TrimSuffix(strings.TrimSpace(interaction.CustomID), "verification")
			if err := m.flowService.CheckPrompt(ctx, interaction.Actor.GuildID, interaction.Actor.UserID, ""); err != nil {
				return responder.Reply(ctx, ephemeralMessage(verificationFlowErrorFromError(err)))
			}
			return responder.ShowModal(ctx, legacyVerificationModal(answer))
		}
		stateID, err := verificationStateIDFromComponent(interaction.CustomID)
		if err != nil {
			return responder.Reply(ctx, ephemeralMessage(verificationFlowErrorMessage("很抱歉，出現了未知的錯誤，請重試!")))
		}
		if err := m.flowService.CheckPrompt(ctx, interaction.Actor.GuildID, interaction.Actor.UserID, stateID); err != nil {
			return responder.Reply(ctx, ephemeralMessage(verificationFlowErrorFromError(err)))
		}
		modal, err := versionedVerificationModal(stateID)
		if err != nil {
			return responder.Reply(ctx, ephemeralMessage(verificationFlowErrorMessage("很抱歉，出現了未知的錯誤，請重試!")))
		}
		return responder.ShowModal(ctx, modal)
	}
}

func (m Module) VerificationAnswerHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		username := interaction.Actor.Username
		if strings.TrimSpace(username) == "" {
			username = strings.Split(interaction.Actor.UserTag, "#")[0]
		}
		var err error
		if interaction.RouteKey.Legacy {
			expected := legacyVerificationExpectedAnswer(interaction)
			answer := firstModalValue(interaction)
			err = m.flowService.CompleteLegacy(ctx, interaction.Actor.GuildID, interaction.Actor.UserID, expected, answer, username)
		} else {
			stateID, parseErr := verificationStateIDFromModal(interaction.CustomID)
			if parseErr != nil {
				err = parseErr
			} else {
				err = m.flowService.Complete(ctx, interaction.Actor.GuildID, interaction.Actor.UserID, stateID, firstModalValue(interaction), username)
			}
		}
		if err != nil {
			return responder.EditOriginal(ctx, verificationAnswerErrorFromError(err))
		}
		return responder.EditOriginal(ctx, verificationFlowSuccessMessage())
	}
}

func verificationPromptMessage(result coreservice.VerificationStartResult) (responses.Message, error) {
	payload, err := customid.StateIDPayload(result.Challenge.StateID)
	if err != nil {
		return responses.Message{}, err
	}
	id, err := customid.Encode(customid.InteractionKindComponent, "verification", "prompt", payload)
	if err != nil {
		return responses.Message{}, err
	}
	return responses.Message{
		Files: []responses.File{{
			Name:        result.ImageName,
			ContentType: "image/jpeg",
			Data:        result.ImageData,
		}},
		Components: []responses.ComponentRow{{
			Components: []responses.Component{{
				Type:     responses.ComponentTypeButton,
				CustomID: id,
				Label:    "點我進行驗證!",
				Emoji:    "<a:arrow:986268851786375218>",
				Style:    responses.ButtonStyleSuccess,
			}},
		}},
		AllowedMentions: &responses.AllowedMentions{},
		Ephemeral:       true,
	}, nil
}

func legacyVerificationModal(answer string) responses.Modal {
	id := strings.TrimSpace(answer) + "ver"
	return responses.Modal{
		CustomID: id,
		Title:    "請輸入驗證碼!",
		Rows: []responses.ModalRow{{
			Inputs: []responses.TextInput{{
				CustomID: id,
				Label:    "請輸入圖片上的驗證碼",
				Style:    responses.TextInputStyleShort,
				Required: true,
			}},
		}},
	}
}

func versionedVerificationModal(stateID string) (responses.Modal, error) {
	payload, err := customid.StateIDPayload(stateID)
	if err != nil {
		return responses.Modal{}, err
	}
	id, err := customid.Encode(customid.InteractionKindModal, "verification", "answer", payload)
	if err != nil {
		return responses.Modal{}, err
	}
	return responses.Modal{
		CustomID: id,
		Title:    "請輸入驗證碼!",
		Rows: []responses.ModalRow{{
			Inputs: []responses.TextInput{{
				CustomID: verificationAnswerInputID,
				Label:    "請輸入圖片上的驗證碼",
				Style:    responses.TextInputStyleShort,
				Required: true,
			}},
		}},
	}, nil
}

func verificationFlowSuccessMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:green_tick:994529015652163614> | 驗證成功，成功給予你身分組及改名(有的話)!",
			Color: joinRoleSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func verificationFlowErrorFromError(err error) responses.Message {
	return verificationFlowErrorMessage(verificationFlowErrorContent(err))
}

func verificationAnswerErrorFromError(err error) responses.Message {
	return verificationAnswerErrorMessage(verificationFlowErrorContent(err))
}

func verificationFlowErrorContent(err error) string {
	switch {
	case errors.Is(err, ports.ErrVerificationConfigMissing):
		return "這服的管理員沒有設置驗證系統，所以不能使用喔!"
	case errors.Is(err, ports.ErrDiscordRoleMissing):
		return "驗證身分組已經不存在了，請通管理員!"
	case errors.Is(err, ports.ErrDiscordRoleNotAssignable):
		return "請通知群主管裡員我沒有權限給你這個身分組(請把我的身分組調高)!"
	case errors.Is(err, coreservice.ErrVerificationAnswerMismatch):
		return "你的驗證碼輸入錯誤，請重試(如果看不清楚的話可以重打指令)"
	case errors.Is(err, coreservice.ErrVerificationOwnerNickname):
		return "你是伺服器服主，我沒有權限改你的名字!"
	case errors.Is(err, domain.ErrInvalidVerificationChallenge), errors.Is(err, domain.ErrInvalidVerificationConfig):
		return "很抱歉，出現了未知的錯誤，請重試!"
	default:
		return "很抱歉，出現了未知的錯誤，請重試!"
	}
}

func verificationFlowErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: joinRoleErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func verificationAnswerErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: content,
			Color: joinRoleErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func ephemeralMessage(message responses.Message) responses.Message {
	message.Ephemeral = true
	return message
}

func verificationStateIDFromComponent(raw string) (string, error) {
	parsed, err := customid.ParseComponent(raw)
	if err != nil {
		return "", err
	}
	return parsed.Payload.StateID, nil
}

func verificationStateIDFromModal(raw string) (string, error) {
	parsed, err := customid.ParseModal(raw, nil)
	if err != nil {
		return "", err
	}
	return parsed.Payload.StateID, nil
}

func legacyVerificationExpectedAnswer(interaction interactions.Interaction) string {
	if len(interaction.ModalFields) > 0 {
		return strings.TrimSuffix(strings.TrimSpace(interaction.ModalFields[0].CustomID), "ver")
	}
	return strings.TrimSuffix(strings.TrimSpace(interaction.CustomID), "ver")
}

func firstModalValue(interaction interactions.Interaction) string {
	for _, field := range interaction.ModalFields {
		if field.CustomID == verificationAnswerInputID || strings.HasSuffix(field.CustomID, "ver") || field.Value != "" {
			return field.Value
		}
	}
	return ""
}
