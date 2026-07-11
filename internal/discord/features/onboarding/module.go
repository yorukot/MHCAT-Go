package onboarding

import (
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/onboarding"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
)

type Module struct {
	service             coreservice.JoinRoleService
	joinRoleConfigured  bool
	leaveMessageService coreservice.LeaveMessageService
	leaveConfigured     bool
	assignmentService   coreservice.JoinRoleAssignmentService
	assignmentEnabled   bool
	welcomeService      coreservice.WelcomeMessageDeliveryService
	welcomeEnabled      bool
	deliveryService     coreservice.LeaveMessageDeliveryService
	deliveryEnabled     bool
	verificationService coreservice.VerificationConfigService
	verificationEnabled bool
	flowService         coreservice.VerificationFlowService
	flowEnabled         bool
	accountAgeService   coreservice.AccountAgeConfigService
	accountAgeEnabled   bool
	accountAgePolicy    coreservice.AccountAgePolicyService
	accountAgeGate      bool
}

func NewModule(repo ports.JoinRoleConfigRepository, roles ports.DiscordRoleInspector) Module {
	return Module{
		service:            coreservice.JoinRoleService{Repository: repo, RoleInspector: roles},
		joinRoleConfigured: true,
	}
}

func NewMessageModule(repo ports.LeaveMessageConfigRepository) Module {
	return Module{
		leaveMessageService: coreservice.LeaveMessageService{Repository: repo},
		leaveConfigured:     true,
	}
}

func NewJoinRoleAssignmentModule(repo ports.JoinRoleConfigReader, roles ports.DiscordRolePort, inspector ports.DiscordRoleInspector, guilds ports.DiscordInfoProvider, directMessages ports.DiscordDirectMessagePort) Module {
	return Module{
		assignmentService: coreservice.JoinRoleAssignmentService{
			Repository:     repo,
			Roles:          roles,
			RoleInspector:  inspector,
			Guilds:         guilds,
			DirectMessages: directMessages,
		},
		assignmentEnabled: repo != nil && roles != nil,
	}
}

func NewWelcomeMessageDeliveryModule(repo ports.JoinMessageConfigReader, messages ports.DiscordMessagePort, channels ports.DiscordCachedChannelReader, special coreservice.SpecialWelcomeConfig) Module {
	return Module{
		welcomeService: coreservice.WelcomeMessageDeliveryService{Repository: repo, Messages: messages, Channels: channels, Special: special},
		welcomeEnabled: repo != nil && messages != nil && channels != nil,
	}
}

func NewLeaveMessageDeliveryModule(repo ports.LeaveMessageConfigReader, messages ports.DiscordMessagePort, channels ports.DiscordCachedChannelReader) Module {
	return Module{
		deliveryService: coreservice.LeaveMessageDeliveryService{Repository: repo, Messages: messages, Channels: channels},
		deliveryEnabled: repo != nil && messages != nil && channels != nil,
	}
}

func NewVerificationModule(repo ports.VerificationConfigRepository, roles ports.DiscordRoleInspector) Module {
	return Module{
		verificationService: coreservice.VerificationConfigService{Repository: repo, RoleInspector: roles},
		verificationEnabled: repo != nil && roles != nil,
	}
}

func NewVerificationFlowModule(repo ports.VerificationConfigReader, roles ports.DiscordRolePort, members ports.DiscordMemberPort, inspector ports.DiscordRoleInspector, guilds ports.DiscordInfoProvider) Module {
	return Module{
		flowService: coreservice.VerificationFlowService{
			Repository: repo,
			Store:      newVerificationChallengeStore(),
			Generator:  verificationCaptchaGenerator{},
			Roles:      roles,
			Members:    members,
			RolesCheck: inspector,
			Guilds:     guilds,
		},
		flowEnabled: repo != nil && roles != nil,
	}
}

func NewAccountAgeModule(repo ports.AccountAgeConfigRepository) Module {
	return Module{
		accountAgeService: coreservice.AccountAgeConfigService{Repository: repo},
		accountAgeEnabled: repo != nil,
	}
}

func NewAccountAgePolicyModule(repo ports.AccountAgeConfigReader, direct ports.DiscordDirectMessagePort, members ports.DiscordMemberPort, messages ports.DiscordMessagePort, channels ports.DiscordChannelPort, guilds ports.DiscordInfoProvider, clock ports.Clock) Module {
	return Module{
		accountAgePolicy: coreservice.AccountAgePolicyService{
			Repository:     repo,
			DirectMessages: direct,
			Members:        members,
			Messages:       messages,
			Channels:       channels,
			Guilds:         guilds,
			Clock:          clock,
		},
		accountAgeGate: repo != nil && members != nil,
	}
}

func (m Module) Name() string {
	if m.accountAgeGate && !m.accountAgeEnabled {
		return "account-age-policy"
	}
	if m.welcomeEnabled && !m.leaveConfigured && !m.joinRoleConfigured {
		return "welcome-message-delivery"
	}
	if m.accountAgeEnabled {
		return "account-age-config"
	}
	if m.flowEnabled && !m.verificationEnabled && !m.joinRoleConfigured && !m.leaveConfigured {
		return "verification-flow"
	}
	if m.verificationEnabled && !m.joinRoleConfigured && !m.leaveConfigured {
		return "verification-config"
	}
	if m.leaveConfigured && !m.joinRoleConfigured {
		return "welcome-message-config"
	}
	return "join-role-config"
}

func (m Module) Commands() []commands.Definition {
	var definitions []commands.Definition
	if m.joinRoleConfigured {
		definitions = append(definitions, JoinRoleDefinitions()...)
	}
	if m.leaveConfigured {
		definitions = append(definitions, MessageDefinitions()...)
	}
	if m.verificationEnabled {
		definitions = append(definitions, VerificationDefinitions()...)
	}
	if m.flowEnabled {
		definitions = append(definitions, VerificationFlowDefinitions()...)
	}
	if m.accountAgeEnabled {
		definitions = append(definitions, AccountAgeDefinitions()...)
	}
	return definitions
}

func (m Module) RegisterRoutes(router *interactions.Router) error {
	if m.joinRoleConfigured {
		if err := router.RegisterSlash(JoinRoleSetCommandName, m.SetHandler()); err != nil {
			return err
		}
		if err := router.RegisterSlash(JoinRoleDeleteCommandName, m.DeleteHandler()); err != nil {
			return err
		}
	}
	if m.leaveConfigured {
		if err := router.RegisterSlash(JoinMessageSetCommandName, m.JoinMessageDashboardHandler()); err != nil {
			return err
		}
		if err := router.RegisterSlash(LeaveMessageSetCommandName, m.LeaveMessagePromptHandler()); err != nil {
			return err
		}
		return router.RegisterRoute(interactions.RouteKey{
			Kind:    interactions.TypeModal,
			Version: "legacy",
			Feature: "welcome",
			Action:  "leave_submit",
			Legacy:  true,
		}, m.LeaveMessageModalHandler())
	}
	if m.verificationEnabled {
		if err := router.RegisterSlash(VerificationSetCommandName, m.VerificationSetHandler()); err != nil {
			return err
		}
	}
	if m.flowEnabled {
		if err := router.RegisterSlash(VerificationCommandName, m.VerificationHandler()); err != nil {
			return err
		}
		for _, key := range []interactions.RouteKey{
			{Kind: interactions.TypeComponent, Version: "v1", Feature: "verification", Action: "prompt"},
			{Kind: interactions.TypeComponent, Version: "legacy", Feature: "verification", Action: "prompt", Legacy: true},
			{Kind: interactions.TypeModal, Version: "v1", Feature: "verification", Action: "answer"},
			{Kind: interactions.TypeModal, Version: "legacy", Feature: "verification", Action: "answer", Legacy: true},
		} {
			handler := m.VerificationPromptHandler()
			if key.Kind == interactions.TypeModal {
				handler = m.VerificationAnswerHandler()
			}
			if err := router.RegisterRoute(key, handler); err != nil {
				return err
			}
		}
	}
	if m.accountAgeEnabled {
		if err := router.RegisterSlash(AccountAgeCommandName, m.AccountAgeHandler()); err != nil {
			return err
		}
	}
	return nil
}

func (m Module) RegisterEventRoutes(dispatcher *events.Dispatcher) {
	if m.accountAgeGate && dispatcher != nil {
		dispatcher.Register(events.TypeMemberAdd, m.AccountAgeGateHandler())
	}
	if m.welcomeEnabled && dispatcher != nil {
		dispatcher.Register(events.TypeMemberAdd, m.WelcomeMessageDeliveryHandler())
	}
	if m.assignmentEnabled && dispatcher != nil {
		dispatcher.Register(events.TypeMemberAdd, m.JoinRoleAssignmentHandler())
	}
	if m.deliveryEnabled && dispatcher != nil {
		dispatcher.Register(events.TypeMemberRemove, m.LeaveMessageDeliveryHandler())
	}
}
