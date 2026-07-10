package onboarding

import (
	"reflect"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/onboarding"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
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
	content := "驗證身分組已經不存在了，請通管理員!"
	wantFlowError := responses.Message{
		Embeds:          []responses.Embed{{Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content, Color: joinRoleErrorColor}},
		AllowedMentions: &responses.AllowedMentions{},
	}
	if got := verificationFlowErrorFromError(ports.ErrDiscordRoleMissing); !reflect.DeepEqual(got, wantFlowError) {
		t.Fatalf("slash/button error = %#v, want %#v", got, wantFlowError)
	}
	wantAnswerError := responses.Message{
		Embeds:          []responses.Embed{{Title: content, Color: joinRoleErrorColor}},
		AllowedMentions: &responses.AllowedMentions{},
	}
	if got := verificationAnswerErrorFromError(ports.ErrDiscordRoleMissing); !reflect.DeepEqual(got, wantAnswerError) {
		t.Fatalf("modal error = %#v, want %#v", got, wantAnswerError)
	}
}
