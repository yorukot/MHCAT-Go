package onboarding

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/events"
)

func (m Module) JoinRoleAssignmentHandler() events.Handler {
	return func(ctx context.Context, event events.Event) error {
		if event.Type != events.TypeMemberAdd {
			return nil
		}
		member := event.Member
		userID := strings.TrimSpace(event.UserID)
		isBot := event.IsBot
		if member != nil {
			if member.UserID != "" {
				userID = member.UserID
			}
			isBot = member.IsBot
		}
		if strings.TrimSpace(event.GuildID) == "" || userID == "" {
			return nil
		}
		err := m.assignmentService.AssignOnJoin(ctx, event.GuildID, userID, isBot)
		if err != nil && ctx.Err() == nil {
			return events.ContinueOnError(err)
		}
		return err
	}
}
