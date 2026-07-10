package onboarding

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

func TestVerificationFlowStartCreatesChallenge(t *testing.T) {
	repo := &fakeVerificationConfigReader{config: domain.VerificationConfig{GuildID: "guild", RoleID: "role"}}
	store := &fakeVerificationChallengeStore{}
	generator := fakeVerificationGenerator{answer: "1234", image: []byte("jpeg")}
	service := VerificationFlowService{Repository: repo, Store: store, Generator: generator, RolesCheck: fakeVerificationRoleInspectorFlow{assignable: true}}

	result, err := service.Start(context.Background(), "guild", "user")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if result.Challenge.StateID == "" || result.Challenge.Answer != "1234" || result.ImageName != "captcha.jpeg" || string(result.ImageData) != "jpeg" {
		t.Fatalf("result = %#v", result)
	}
}

func TestVerificationFlowCompleteAddsRoleAndNickname(t *testing.T) {
	repo := &fakeVerificationConfigReader{config: domain.VerificationConfig{GuildID: "guild", RoleID: "role", RenameTemplate: "{name} | MHCAT"}}
	store := &fakeVerificationChallengeStore{challenge: domain.VerificationChallenge{StateID: "state", GuildID: "guild", UserID: "user", Answer: "1234"}}
	roles := &fakeVerificationRolePort{}
	members := &fakeVerificationMemberPort{}
	guilds := &fakeVerificationGuildInfo{ownerID: "owner"}
	service := VerificationFlowService{Repository: repo, Store: store, Roles: roles, Members: members, RolesCheck: fakeVerificationRoleInspectorFlow{assignable: true}, Guilds: guilds}

	err := service.Complete(context.Background(), "guild", "user", "state", "1234", "Yoru")
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if roles.added != "guild:user:role" {
		t.Fatalf("role add = %q", roles.added)
	}
	if members.nickname != "guild:user:Yoru | MHCAT" {
		t.Fatalf("nickname = %q", members.nickname)
	}
	if !store.deleted {
		t.Fatal("challenge should be deleted after success")
	}
}

func TestVerificationFlowTreatsWhitespaceRenameAsConfigured(t *testing.T) {
	repo := &fakeVerificationConfigReader{config: domain.VerificationConfig{GuildID: "guild", RoleID: "role", RenameTemplate: "  "}}
	store := &fakeVerificationChallengeStore{challenge: domain.VerificationChallenge{StateID: "state", GuildID: "guild", UserID: "user", Answer: "1234"}}
	members := &fakeVerificationMemberPort{}
	service := VerificationFlowService{
		Repository: repo,
		Store:      store,
		Roles:      &fakeVerificationRolePort{},
		Members:    members,
		RolesCheck: fakeVerificationRoleInspectorFlow{assignable: true},
		Guilds:     &fakeVerificationGuildInfo{ownerID: "owner"},
	}
	if err := service.Complete(context.Background(), "guild", "user", "state", "1234", "Yoru"); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if members.nickname != "guild:user:  " {
		t.Fatalf("nickname = %q", members.nickname)
	}
}

func TestVerificationFlowCompleteRejectsWrongAnswerAndOwnerNickname(t *testing.T) {
	repo := &fakeVerificationConfigReader{config: domain.VerificationConfig{GuildID: "guild", RoleID: "role", RenameTemplate: "{name}"}}
	store := &fakeVerificationChallengeStore{challenge: domain.VerificationChallenge{StateID: "state", GuildID: "guild", UserID: "user", Answer: "1234"}}
	service := VerificationFlowService{Repository: repo, Store: store, Roles: &fakeVerificationRolePort{}, Members: &fakeVerificationMemberPort{}, RolesCheck: fakeVerificationRoleInspectorFlow{assignable: true}, Guilds: &fakeVerificationGuildInfo{ownerID: "user"}}

	if err := service.Complete(context.Background(), "guild", "user", "state", "9999", "Yoru"); !errors.Is(err, ErrVerificationAnswerMismatch) {
		t.Fatalf("wrong answer error = %v", err)
	}
	if err := service.Complete(context.Background(), "guild", "user", "state", " 1234 ", "Yoru"); !errors.Is(err, ErrVerificationAnswerMismatch) {
		t.Fatalf("whitespace answer error = %v", err)
	}
	if err := service.Complete(context.Background(), "guild", "user", "state", "1234", "Yoru"); !errors.Is(err, ErrVerificationOwnerNickname) {
		t.Fatalf("owner nickname error = %v", err)
	}
}

func TestVerificationFlowLegacyAnswerComparisonPreservesWhitespace(t *testing.T) {
	service := VerificationFlowService{}
	if err := service.CompleteLegacy(context.Background(), "guild", "user", "AB12", " AB12 ", "Yoru"); !errors.Is(err, ErrVerificationAnswerMismatch) {
		t.Fatalf("whitespace answer error = %v", err)
	}
}

func TestVerificationFlowCheckPromptValidatesStateAndRole(t *testing.T) {
	repo := &fakeVerificationConfigReader{config: domain.VerificationConfig{GuildID: "guild", RoleID: "role"}}
	store := &fakeVerificationChallengeStore{challenge: domain.VerificationChallenge{StateID: "state", GuildID: "guild", UserID: "user", Answer: "1234"}}
	service := VerificationFlowService{Repository: repo, Store: store, RolesCheck: fakeVerificationRoleInspectorFlow{assignable: true}}
	if err := service.CheckPrompt(context.Background(), "guild", "user", "state"); err != nil {
		t.Fatalf("check prompt: %v", err)
	}
	if err := service.CheckPrompt(context.Background(), "guild", "other-user", "state"); !errors.Is(err, ErrVerificationAnswerMismatch) {
		t.Fatalf("mismatched state owner error = %v", err)
	}

	service.RolesCheck = fakeVerificationRoleInspectorFlow{assignable: false}
	if err := service.CheckPrompt(context.Background(), "guild", "user", "state"); !errors.Is(err, ports.ErrDiscordRoleNotAssignable) {
		t.Fatalf("unassignable role error = %v", err)
	}
}

type fakeVerificationConfigReader struct {
	config domain.VerificationConfig
	err    error
}

func (r *fakeVerificationConfigReader) GetVerificationConfig(context.Context, string) (domain.VerificationConfig, error) {
	if r.err != nil {
		return domain.VerificationConfig{}, r.err
	}
	return r.config, nil
}

type fakeVerificationChallengeStore struct {
	challenge domain.VerificationChallenge
	deleted   bool
}

func (s *fakeVerificationChallengeStore) Create(_ context.Context, challenge domain.VerificationChallenge) (domain.VerificationChallenge, error) {
	challenge.StateID = "state"
	s.challenge = challenge
	return challenge, nil
}

func (s *fakeVerificationChallengeStore) Get(context.Context, string) (domain.VerificationChallenge, error) {
	return s.challenge, nil
}

func (s *fakeVerificationChallengeStore) Delete(context.Context, string) error {
	s.deleted = true
	return nil
}

type fakeVerificationGenerator struct {
	answer string
	image  []byte
}

func (g fakeVerificationGenerator) Generate(context.Context) (VerificationGeneratedChallenge, error) {
	return VerificationGeneratedChallenge{Answer: g.answer, ImageName: "captcha.jpeg", ImageData: g.image}, nil
}

type fakeVerificationRoleInspectorFlow struct {
	assignable bool
	err        error
}

func (i fakeVerificationRoleInspectorFlow) CanAssignRole(context.Context, string, string) (bool, error) {
	return i.assignable, i.err
}

type fakeVerificationRolePort struct {
	added string
}

func (p *fakeVerificationRolePort) AddRole(_ context.Context, guildID string, userID string, roleID string) error {
	p.added = guildID + ":" + userID + ":" + roleID
	return nil
}

func (p *fakeVerificationRolePort) RemoveRole(context.Context, string, string, string) error {
	return nil
}

type fakeVerificationMemberPort struct {
	nickname string
}

func (p *fakeVerificationMemberPort) MoveMember(context.Context, string, string, *string) error {
	return nil
}

func (p *fakeVerificationMemberPort) SetNickname(_ context.Context, guildID string, userID string, nickname string) error {
	p.nickname = guildID + ":" + userID + ":" + nickname
	return nil
}

func (p *fakeVerificationMemberPort) KickMember(context.Context, string, string, string) error {
	return nil
}

func (p *fakeVerificationMemberPort) BanMember(context.Context, string, string, string, int) error {
	return nil
}

type fakeVerificationGuildInfo struct {
	ownerID string
}

func (p *fakeVerificationGuildInfo) UserInfo(context.Context, string, string) (ports.DiscordUserInfo, error) {
	return ports.DiscordUserInfo{}, nil
}

func (p *fakeVerificationGuildInfo) GuildInfo(context.Context, string) (ports.DiscordGuildInfo, error) {
	return ports.DiscordGuildInfo{OwnerID: p.ownerID}, nil
}
