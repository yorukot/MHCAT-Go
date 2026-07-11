package onboarding

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/onboarding"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
)

func TestVerificationSetupMessagesPreserveLegacyPayloads(t *testing.T) {
	tests := []struct {
		name string
		got  responses.Message
		want responses.Message
	}{
		{
			name: "success with raw rename",
			got:  verificationSuccessMessage("role-1", "  {name} | MHCAT  "),
			want: responses.Message{
				Embeds: []responses.Embed{{
					Title:       "<a:green_tick:994529015652163614> 設置成功!",
					Description: "<:roleplaying:985945121264635964>身分組: <@&role-1>!\n <:id:985950321975128094>改名為:  {name} | MHCAT  ",
					Color:       joinRoleSuccessColor,
				}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "success without rename",
			got:  verificationSuccessMessage("role-1", ""),
			want: responses.Message{
				Embeds: []responses.Embed{{
					Title:       "<a:green_tick:994529015652163614> 設置成功!",
					Description: "<:roleplaying:985945121264635964>身分組: <@&role-1>!\n <:id:985950321975128094>改名為:null",
					Color:       joinRoleSuccessColor,
				}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "permission error",
			got:  verificationErrorMessage("你需要有`訊息管理`才能使用此指令"),
			want: responses.Message{
				Embeds:          []responses.Embed{{Title: "<a:Discord_AnimatedNo:1015989839809757295> | 你需要有`訊息管理`才能使用此指令", Color: joinRoleErrorColor}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "role hierarchy error",
			got:  verificationErrorFromError(ports.ErrDiscordRoleNotAssignable),
			want: responses.Message{
				Embeds:          []responses.Embed{{Title: "<a:Discord_AnimatedNo:1015989839809757295> | 我沒有權限為大家增加這個身分組，請將我的身分組位階調高", Color: joinRoleErrorColor}},
				AllowedMentions: &responses.AllowedMentions{},
			},
		},
		{
			name: "unknown setup error",
			got:  verificationErrorFromError(errors.New("boom")),
			want: responses.Message{
				Embeds:          []responses.Embed{{Title: "<a:Discord_AnimatedNo:1015989839809757295> | 很抱歉，出現了未知的錯誤，請重試!", Color: joinRoleErrorColor}},
				AllowedMentions: &responses.AllowedMentions{},
			},
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

func TestVerificationPromptMessagePreservesLegacyVisiblePayload(t *testing.T) {
	got, err := verificationPromptMessage(coreservice.VerificationStartResult{
		Challenge: domain.VerificationChallenge{StateID: "state123", GuildID: "guild-1", UserID: "user-1", Answer: "ABCDEF"},
		ImageName: "captcha.jpeg",
		ImageData: []byte("jpeg"),
	})
	if err != nil {
		t.Fatalf("prompt message: %v", err)
	}
	want := responses.Message{
		Files: []responses.File{{Name: "captcha.jpeg", ContentType: "image/jpeg", Data: []byte("jpeg")}},
		Components: []responses.ComponentRow{{Components: []responses.Component{{
			Type:     responses.ComponentTypeButton,
			CustomID: "mhcat:v1:verification:prompt:state=state123",
			Label:    "點我進行驗證!",
			Emoji:    "<a:arrow:986268851786375218>",
			Style:    responses.ButtonStyleSuccess,
		}}}},
		AllowedMentions: &responses.AllowedMentions{},
		Ephemeral:       true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("message = %#v, want %#v", got, want)
	}
}

func TestVerificationModalsPreserveLegacyVisiblePayload(t *testing.T) {
	wantLegacy := responses.Modal{
		CustomID: "ABCDEFver",
		Title:    "請輸入驗證碼!",
		Rows: []responses.ModalRow{{Inputs: []responses.TextInput{{
			CustomID: "ABCDEFver",
			Label:    "請輸入圖片上的驗證碼",
			Style:    responses.TextInputStyleShort,
			Required: true,
		}}}},
	}
	if got := legacyVerificationModal("ABCDEF"); !reflect.DeepEqual(got, wantLegacy) {
		t.Fatalf("legacy modal = %#v, want %#v", got, wantLegacy)
	}

	gotVersioned, err := versionedVerificationModal("state123")
	if err != nil {
		t.Fatalf("versioned modal: %v", err)
	}
	wantVersioned := responses.Modal{
		CustomID: "mhcat:v1:verification:answer:state=state123",
		Title:    "請輸入驗證碼!",
		Rows: []responses.ModalRow{{Inputs: []responses.TextInput{{
			CustomID: verificationAnswerInputID,
			Label:    "請輸入圖片上的驗證碼",
			Style:    responses.TextInputStyleShort,
			Required: true,
		}}}},
	}
	if !reflect.DeepEqual(gotVersioned, wantVersioned) {
		t.Fatalf("versioned modal = %#v, want %#v", gotVersioned, wantVersioned)
	}
}

func TestVerificationCompletionMessagesPreserveContextSpecificPayloads(t *testing.T) {
	wantSuccess := responses.Message{
		Embeds:          []responses.Embed{{Title: "<a:green_tick:994529015652163614> | 驗證成功，成功給予你身分組及改名(有的話)!", Color: joinRoleSuccessColor}},
		AllowedMentions: &responses.AllowedMentions{},
	}
	if got := verificationFlowSuccessMessage(); !reflect.DeepEqual(got, wantSuccess) {
		t.Fatalf("success = %#v, want %#v", got, wantSuccess)
	}
	tests := []struct {
		name    string
		err     error
		content string
	}{
		{name: "missing config", err: ports.ErrVerificationConfigMissing, content: "這服的管理員沒有設置驗證系統，所以不能使用喔!"},
		{name: "missing role", err: ports.ErrDiscordRoleMissing, content: "驗證身分組已經不存在了，請通管理員!"},
		{name: "role hierarchy", err: ports.ErrDiscordRoleNotAssignable, content: "請通知群主管裡員我沒有權限給你這個身分組(請把我的身分組調高)!"},
		{name: "wrong answer", err: coreservice.ErrVerificationAnswerMismatch, content: "你的驗證碼輸入錯誤，請重試(如果看不清楚的話可以重打指令)"},
		{name: "owner rename", err: coreservice.ErrVerificationOwnerNickname, content: "你是伺服器服主，我沒有權限改你的名字!"},
		{name: "invalid state", err: domain.ErrInvalidVerificationChallenge, content: "很抱歉，出現了未知的錯誤，請重試!"},
		{name: "unknown", err: errors.New("boom"), content: "很抱歉，出現了未知的錯誤，請重試!"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wantFlowError := responses.Message{
				Embeds:          []responses.Embed{{Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + test.content, Color: joinRoleErrorColor}},
				AllowedMentions: &responses.AllowedMentions{},
			}
			if got := verificationFlowErrorFromError(test.err); !reflect.DeepEqual(got, wantFlowError) {
				t.Fatalf("slash/button error = %#v, want %#v", got, wantFlowError)
			}
			wantAnswerError := responses.Message{
				Embeds:          []responses.Embed{{Title: test.content, Color: joinRoleErrorColor}},
				AllowedMentions: &responses.AllowedMentions{},
			}
			if got := verificationAnswerErrorFromError(test.err); !reflect.DeepEqual(got, wantAnswerError) {
				t.Fatalf("modal error = %#v, want %#v", got, wantAnswerError)
			}
		})
	}
}

func TestVerificationHandlersPreserveLegacyResponseVisibility(t *testing.T) {
	configModule := NewVerificationModule(nil, nil)
	configResponder := fakediscord.NewResponder()
	if err := configModule.VerificationSetHandler()(context.Background(), fakediscord.SlashInteraction(VerificationSetCommandName), configResponder); err != nil {
		t.Fatalf("config handler: %v", err)
	}
	if len(configResponder.Defers) != 1 || configResponder.Defers[0].Ephemeral {
		t.Fatalf("config defer should be public: %#v", configResponder.Defers)
	}

	flowModule, _, sideEffects := newVerificationFlowTestModule()
	sideEffects.AssignableRoles["guild-1/role-1"] = true
	flowResponder := fakediscord.NewResponder()
	if err := flowModule.VerificationHandler()(context.Background(), fakediscord.SlashInteraction(VerificationCommandName), flowResponder); err != nil {
		t.Fatalf("flow handler: %v", err)
	}
	if len(flowResponder.Defers) != 1 || !flowResponder.Defers[0].Ephemeral {
		t.Fatalf("flow defer should be ephemeral: %#v", flowResponder.Defers)
	}

	answerResponder := fakediscord.NewResponder()
	answer := fakediscord.SlashInteraction(VerificationCommandName)
	answer.RouteKey.Legacy = true
	answer.CustomID = "ABCDEFver"
	if err := flowModule.VerificationAnswerHandler()(context.Background(), answer, answerResponder); err != nil {
		t.Fatalf("answer handler: %v", err)
	}
	if len(answerResponder.Defers) != 1 || answerResponder.Defers[0].Ephemeral {
		t.Fatalf("modal defer should be public: %#v", answerResponder.Defers)
	}
}
