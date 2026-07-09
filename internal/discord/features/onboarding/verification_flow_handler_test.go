package onboarding

import (
	"context"
	"strings"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/onboarding"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakebotinfo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakediscord"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakeusage"
)

func TestVerificationHandlerShowsLegacyPromptWithSafeStateID(t *testing.T) {
	module, _, sideEffects, usage := newVerificationFlowTestModule()
	interaction := fakediscord.SlashInteraction(VerificationCommandName)
	responder := fakediscord.NewResponder()
	sideEffects.AssignableRoles["guild-1/role-1"] = true

	if err := module.VerificationHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if len(responder.Defers) != 1 || !responder.Defers[0].Ephemeral {
		t.Fatalf("defers = %#v", responder.Defers)
	}
	if len(responder.Edits) != 1 {
		t.Fatalf("edits = %#v", responder.Edits)
	}
	edit := responder.Edits[0]
	if len(edit.Files) != 1 || edit.Files[0].Name != "captcha.jpeg" || edit.Files[0].ContentType != "image/jpeg" || len(edit.Files[0].Data) == 0 {
		t.Fatalf("files = %#v", edit.Files)
	}
	button := edit.Components[0].Components[0]
	if button.Label != "點我進行驗證!" || button.Emoji != "<a:arrow:986268851786375218>" || button.Style != "success" {
		t.Fatalf("button = %#v", button)
	}
	if !strings.HasPrefix(button.CustomID, "mhcat:v1:verification:prompt:state=") || strings.Contains(button.CustomID, "1234") {
		t.Fatalf("custom id should use state id, got %q", button.CustomID)
	}
	if len(usage.Events) != 1 || usage.Events[0].CommandName != VerificationCommandName {
		t.Fatalf("usage = %#v", usage.Events)
	}
}

func TestVerificationPromptShowsVersionedAndLegacyModal(t *testing.T) {
	module, _, sideEffects, _ := newVerificationFlowTestModule()
	sideEffects.AssignableRoles["guild-1/role-1"] = true
	start := fakediscord.NewResponder()
	if err := module.VerificationHandler()(context.Background(), fakediscord.SlashInteraction(VerificationCommandName), start); err != nil {
		t.Fatalf("start: %v", err)
	}
	interaction := fakediscord.ComponentInteractionFromID(start.Edits[0].Components[0].Components[0].CustomID)
	interaction.RouteKey = interactions.RouteKey{Kind: interactions.TypeComponent, Version: "v1", Feature: "verification", Action: "prompt"}
	responder := fakediscord.NewResponder()
	if err := module.VerificationPromptHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("prompt: %v", err)
	}
	if len(responder.Modals) != 1 || responder.Modals[0].Title != "請輸入驗證碼!" || responder.Modals[0].Rows[0].Inputs[0].Label != "請輸入圖片上的驗證碼" {
		t.Fatalf("modal = %#v", responder.Modals)
	}
	if !strings.HasPrefix(responder.Modals[0].CustomID, "mhcat:v1:verification:answer:state=") {
		t.Fatalf("modal custom id = %q", responder.Modals[0].CustomID)
	}

	legacy := fakediscord.ComponentInteractionFromID("AB12verification")
	legacy.RouteKey = interactions.RouteKey{Kind: interactions.TypeComponent, Version: "legacy", Feature: "verification", Action: "prompt", Legacy: true}
	responder = fakediscord.NewResponder()
	if err := module.VerificationPromptHandler()(context.Background(), legacy, responder); err != nil {
		t.Fatalf("legacy prompt: %v", err)
	}
	if responder.Modals[0].CustomID != "AB12ver" || responder.Modals[0].Rows[0].Inputs[0].CustomID != "AB12ver" {
		t.Fatalf("legacy modal = %#v", responder.Modals[0])
	}
}

func TestVerificationPromptRechecksRoleBeforeModal(t *testing.T) {
	module, _, sideEffects, _ := newVerificationFlowTestModule()
	sideEffects.AssignableRoles["guild-1/role-1"] = true
	start := fakediscord.NewResponder()
	if err := module.VerificationHandler()(context.Background(), fakediscord.SlashInteraction(VerificationCommandName), start); err != nil {
		t.Fatalf("start: %v", err)
	}
	sideEffects.MissingRoles["guild-1/role-1"] = true
	interaction := fakediscord.ComponentInteractionFromID(start.Edits[0].Components[0].Components[0].CustomID)
	interaction.RouteKey = interactions.RouteKey{Kind: interactions.TypeComponent, Version: "v1", Feature: "verification", Action: "prompt"}
	responder := fakediscord.NewResponder()
	if err := module.VerificationPromptHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("prompt: %v", err)
	}
	if len(responder.Modals) != 0 {
		t.Fatalf("modal should not open when role is missing: %#v", responder.Modals)
	}
	if len(responder.Replies) != 1 || !responder.Replies[0].Ephemeral || !strings.Contains(responder.Replies[0].Embeds[0].Title, "驗證身分組已經不存在") {
		t.Fatalf("reply = %#v", responder.Replies)
	}
}

func TestVerificationAnswerAddsRoleNicknameAndReturnsLegacySuccess(t *testing.T) {
	module, _, sideEffects, _ := newVerificationFlowTestModule()
	sideEffects.AssignableRoles["guild-1/role-1"] = true
	interaction := fakediscord.SlashInteraction(VerificationCommandName)
	responder := fakediscord.NewResponder()
	if err := module.VerificationHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("start: %v", err)
	}
	buttonID := responder.Edits[0].Components[0].Components[0].CustomID
	modal, err := versionedVerificationModal(mustStateIDFromComponent(t, buttonID))
	if err != nil {
		t.Fatal(err)
	}
	answer := interactions.Interaction{
		Type:     interactions.TypeModal,
		CustomID: modal.CustomID,
		RouteKey: interactions.RouteKey{Kind: interactions.TypeModal, Version: "v1", Feature: "verification", Action: "answer"},
		Actor:    interactions.Actor{GuildID: "guild-1", UserID: "user-1", Username: "Yoru", UserTag: "Yoru#0001"},
		ModalFields: []customid.ModalField{{
			CustomID: verificationAnswerInputID,
			Value:    fixedVerificationAnswer,
		}},
	}
	responder = fakediscord.NewResponder()
	if err := module.VerificationAnswerHandler()(context.Background(), answer, responder); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if len(sideEffects.AddedRoles) != 1 || sideEffects.AddedRoles[0].RoleID != "role-1" {
		t.Fatalf("roles = %#v", sideEffects.AddedRoles)
	}
	if len(sideEffects.Nicknames) != 1 || sideEffects.Nicknames[0].Nickname != "Yoru | MHCAT" {
		t.Fatalf("nicknames = %#v", sideEffects.Nicknames)
	}
	if len(responder.Edits) != 1 || responder.Edits[0].Embeds[0].Title != "<a:green_tick:994529015652163614> | 驗證成功，成功給予你身分組及改名(有的話)!" {
		t.Fatalf("success = %#v", responder.Edits)
	}
}

func TestVerificationLegacyAnswerAndWrongAnswer(t *testing.T) {
	module, _, sideEffects, _ := newVerificationFlowTestModule()
	sideEffects.AssignableRoles["guild-1/role-1"] = true
	interaction := interactions.Interaction{
		Type:     interactions.TypeModal,
		CustomID: "AB12ver",
		RouteKey: interactions.RouteKey{Kind: interactions.TypeModal, Version: "legacy", Feature: "verification", Action: "answer", Legacy: true},
		Actor:    interactions.Actor{GuildID: "guild-1", UserID: "user-1", Username: "Yoru"},
		ModalFields: []customid.ModalField{{
			CustomID: "AB12ver",
			Value:    "WRONG",
		}},
	}
	responder := fakediscord.NewResponder()
	if err := module.VerificationAnswerHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("wrong answer handler: %v", err)
	}
	if !strings.Contains(responder.Edits[0].Embeds[0].Title, "你的驗證碼輸入錯誤") {
		t.Fatalf("wrong answer response = %#v", responder.Edits)
	}

	interaction.ModalFields[0].Value = "AB12"
	responder = fakediscord.NewResponder()
	if err := module.VerificationAnswerHandler()(context.Background(), interaction, responder); err != nil {
		t.Fatalf("legacy answer handler: %v", err)
	}
	if len(sideEffects.AddedRoles) != 1 {
		t.Fatalf("roles = %#v", sideEffects.AddedRoles)
	}
}

const fixedVerificationAnswer = "1234"

func newVerificationFlowTestModule() (Module, *fakemongo.VerificationConfigRepository, *fakediscord.SideEffects, *fakeusage.Tracker) {
	repo := fakemongo.NewVerificationConfigRepository()
	repo.Configs["guild-1"] = domain.VerificationConfig{GuildID: "guild-1", RoleID: "role-1", RenameTemplate: "{name} | MHCAT"}
	sideEffects := fakediscord.NewSideEffects()
	usage := &fakeusage.Tracker{}
	guilds := &fakebotinfo.DiscordInfoProvider{Guild: domainToDiscordGuild("owner-1")}
	module := NewVerificationFlowModule(repo, sideEffects, sideEffects, sideEffects, guilds, usage)
	module.flowService.Generator = fixedVerificationGenerator{}
	return module, repo, sideEffects, usage
}

type fixedVerificationGenerator struct{}

func (fixedVerificationGenerator) Generate(context.Context) (coreservice.VerificationGeneratedChallenge, error) {
	return coreservice.VerificationGeneratedChallenge{Answer: fixedVerificationAnswer, ImageName: "captcha.jpeg", ImageData: []byte("jpeg")}, nil
}

func domainToDiscordGuild(ownerID string) ports.DiscordGuildInfo {
	return ports.DiscordGuildInfo{OwnerID: ownerID}
}

func mustStateIDFromComponent(t *testing.T, id string) string {
	t.Helper()
	stateID, err := verificationStateIDFromComponent(id)
	if err != nil {
		t.Fatal(err)
	}
	return stateID
}
