package onboarding

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

var ErrVerificationAnswerMismatch = errors.New("verification answer mismatch")
var ErrVerificationOwnerNickname = errors.New("verification owner nickname change is not allowed")

type VerificationChallengeStore interface {
	Create(ctx context.Context, challenge domain.VerificationChallenge) (domain.VerificationChallenge, error)
	Get(ctx context.Context, stateID string) (domain.VerificationChallenge, error)
	Delete(ctx context.Context, stateID string) error
}

type VerificationChallengeGenerator interface {
	Generate(ctx context.Context) (VerificationGeneratedChallenge, error)
}

type VerificationGeneratedChallenge struct {
	Answer    string
	ImageName string
	ImageData []byte
}

type VerificationStartResult struct {
	Config    domain.VerificationConfig
	Challenge domain.VerificationChallenge
	ImageName string
	ImageData []byte
}

type VerificationFlowService struct {
	Repository ports.VerificationConfigReader
	Store      VerificationChallengeStore
	Generator  VerificationChallengeGenerator
	Roles      ports.DiscordRolePort
	Members    ports.DiscordMemberPort
	RolesCheck ports.DiscordRoleInspector
	Guilds     ports.DiscordInfoProvider
}

func (s VerificationFlowService) Start(ctx context.Context, guildID string, userID string) (VerificationStartResult, error) {
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return VerificationStartResult{}, domain.ErrInvalidVerificationChallenge
	}
	if s.Repository == nil || s.Store == nil || s.Generator == nil {
		return VerificationStartResult{}, domain.ErrInvalidVerificationChallenge
	}
	config, err := s.loadAndCheckConfig(ctx, guildID)
	if err != nil {
		return VerificationStartResult{}, err
	}
	generated, err := s.Generator.Generate(ctx)
	if err != nil {
		return VerificationStartResult{}, err
	}
	challenge, err := s.Store.Create(ctx, domain.VerificationChallenge{
		GuildID: guildID,
		UserID:  userID,
		Answer:  generated.Answer,
	})
	if err != nil {
		return VerificationStartResult{}, err
	}
	if err := challenge.Validate(); err != nil {
		return VerificationStartResult{}, err
	}
	imageName := strings.TrimSpace(generated.ImageName)
	if imageName == "" {
		imageName = "captcha.jpeg"
	}
	return VerificationStartResult{Config: config, Challenge: challenge, ImageName: imageName, ImageData: generated.ImageData}, nil
}

func (s VerificationFlowService) CheckPrompt(ctx context.Context, guildID string, userID string, stateID string) error {
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	stateID = strings.TrimSpace(stateID)
	if guildID == "" {
		return domain.ErrInvalidVerificationChallenge
	}
	if stateID != "" {
		if userID == "" || s.Store == nil {
			return domain.ErrInvalidVerificationChallenge
		}
		challenge, err := s.Store.Get(ctx, stateID)
		if err != nil {
			return err
		}
		if strings.TrimSpace(challenge.GuildID) != guildID || strings.TrimSpace(challenge.UserID) != userID {
			return ErrVerificationAnswerMismatch
		}
	}
	_, err := s.loadAndCheckConfig(ctx, guildID)
	return err
}

func (s VerificationFlowService) Complete(ctx context.Context, guildID string, userID string, stateID string, answer string, username string) error {
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	stateID = strings.TrimSpace(stateID)
	if guildID == "" || userID == "" || stateID == "" || answer == "" {
		return domain.ErrInvalidVerificationChallenge
	}
	if s.Repository == nil || s.Store == nil || s.Roles == nil {
		return domain.ErrInvalidVerificationChallenge
	}
	challenge, err := s.Store.Get(ctx, stateID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(challenge.GuildID) != guildID || strings.TrimSpace(challenge.UserID) != userID || challenge.Answer != answer {
		return ErrVerificationAnswerMismatch
	}
	if err := s.completeConfigured(ctx, guildID, userID, username); err != nil {
		return err
	}
	_ = s.Store.Delete(ctx, stateID)
	return ctx.Err()
}

func (s VerificationFlowService) CompleteLegacy(ctx context.Context, guildID string, userID string, expectedAnswer string, answer string, username string) error {
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" || expectedAnswer == "" || answer == "" {
		return domain.ErrInvalidVerificationChallenge
	}
	if expectedAnswer != answer {
		return ErrVerificationAnswerMismatch
	}
	return s.completeConfigured(ctx, guildID, userID, username)
}

func (s VerificationFlowService) completeConfigured(ctx context.Context, guildID string, userID string, username string) error {
	if s.Repository == nil || s.Roles == nil {
		return domain.ErrInvalidVerificationChallenge
	}
	config, err := s.loadAndCheckConfig(ctx, guildID)
	if err != nil {
		return err
	}
	if err := s.Roles.AddRole(ctx, guildID, userID, config.RoleID); err != nil {
		return err
	}
	if config.RenameTemplate != "" {
		if s.Members == nil || s.Guilds == nil {
			return domain.ErrInvalidVerificationChallenge
		}
		guild, err := s.Guilds.GuildInfo(ctx, guildID)
		if err != nil {
			return err
		}
		if strings.TrimSpace(guild.OwnerID) == userID {
			return ErrVerificationOwnerNickname
		}
		nickname := strings.Replace(config.RenameTemplate, "{name}", strings.TrimSpace(username), 1)
		if err := s.Members.SetNickname(ctx, guildID, userID, nickname); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func (s VerificationFlowService) loadAndCheckConfig(ctx context.Context, guildID string) (domain.VerificationConfig, error) {
	if s.Repository == nil {
		return domain.VerificationConfig{}, domain.ErrInvalidVerificationChallenge
	}
	config, err := s.Repository.GetVerificationConfig(ctx, guildID)
	if err != nil {
		return domain.VerificationConfig{}, err
	}
	if s.RolesCheck != nil {
		ok, err := s.RolesCheck.CanAssignRole(ctx, guildID, config.RoleID)
		if err != nil {
			return domain.VerificationConfig{}, err
		}
		if !ok {
			return domain.VerificationConfig{}, ports.ErrDiscordRoleNotAssignable
		}
	}
	return config, nil
}
