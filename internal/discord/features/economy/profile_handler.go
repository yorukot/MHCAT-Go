package economy

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	coreeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/economy"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	profileFileName           = "user-info.png"
	profileFileContentType    = "image/png"
	profileRefreshEmoji       = "<:update:1020532095212335235>"
	profileRefreshLabel       = "更新"
	profileMissingMemberTitle = "<a:error:980086028113182730> | 很抱歉，這位使用者已退出該伺服器!!"
)

var legacyProfileRefreshRe = regexp.MustCompile(`^([0-9]{17,20})my-profile$`)

func (m Module) ProfileHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.profile.Repository == nil {
			return domain.ErrInvalidEconomyProfileQuery
		}
		if err := responder.Reply(ctx, signLoadingMessage(legacyCoinRankPNGURL(interaction.Actor.AvatarURL))); err != nil {
			return err
		}
		targetUserID := profileTargetUserID(interaction)
		message, err := m.profileMessage(ctx, interaction, targetUserID)
		if err != nil {
			return err
		}
		if err := responder.EditOriginal(ctx, message); err != nil {
			return err
		}
		return m.trackCommand(ctx, interaction, ProfileCommandName)
	}
}

func (m Module) ProfileRefreshHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.profile.Repository == nil {
			return domain.ErrInvalidEconomyProfileQuery
		}
		targetUserID, err := parseLegacyProfileRefresh(interaction.CustomID)
		if err != nil {
			return err
		}
		userInfo, err := m.lookupProfileUser(ctx, interaction, targetUserID)
		if err != nil {
			return responder.Reply(ctx, profileMissingMemberMessage())
		}
		loading := signLoadingMessage(legacyCoinRankPNGURL(interaction.Actor.AvatarURL))
		loading.ClearAttachments = true
		if err := responder.UpdateMessage(ctx, loading); err != nil {
			return err
		}
		message, err := m.profileMessageWithUser(ctx, interaction, targetUserID, userInfo)
		if err != nil {
			return err
		}
		return responder.EditOriginal(ctx, message)
	}
}

func profileTargetUserID(interaction interactions.Interaction) string {
	for _, key := range []string{"user", "使用者"} {
		if value := strings.TrimSpace(interaction.Options[key]); value != "" {
			return value
		}
	}
	return interaction.Actor.UserID
}

func parseLegacyProfileRefresh(raw string) (string, error) {
	matches := legacyProfileRefreshRe.FindStringSubmatch(raw)
	if matches == nil {
		return "", domain.ErrInvalidEconomyProfileQuery
	}
	return matches[1], nil
}

func (m Module) profileMessage(ctx context.Context, interaction interactions.Interaction, targetUserID string) (responses.Message, error) {
	targetUserID = strings.TrimSpace(targetUserID)
	if targetUserID == "" {
		targetUserID = interaction.Actor.UserID
	}
	userInfo, err := m.lookupProfileUser(ctx, interaction, targetUserID)
	if err != nil {
		return responses.Message{}, err
	}
	return m.profileMessageWithUser(ctx, interaction, targetUserID, userInfo)
}

func (m Module) profileMessageWithUser(ctx context.Context, interaction interactions.Interaction, targetUserID string, userInfo ports.DiscordUserInfo) (responses.Message, error) {
	guildInfo := m.lookupCoinRankGuild(ctx, interaction.Actor.GuildID)
	now := clockOrSystem(m.clock).Now()
	result, err := m.profile.Query(ctx, coreeconomy.ProfileQuery{
		GuildID: interaction.Actor.GuildID,
		UserID:  targetUserID,
		Now:     now,
	})
	if err != nil {
		return responses.Message{}, err
	}
	avatarData := fetchProfileAvatar(ctx, userInfo.AvatarURL)
	image, err := renderProfilePNG(profileCanvasView{
		DisplayName:    legacyProfileDisplayName(userInfo, targetUserID),
		GuildName:      profileGuildName(guildInfo, interaction.Actor.GuildID),
		UserCreatedAt:  userInfo.CreatedAt,
		MemberJoinedAt: userInfo.JoinedAt,
		AvatarData:     avatarData,
		Result:         result,
	})
	if err != nil {
		return responses.Message{}, err
	}
	return responses.Message{
		Files: []responses.File{{
			Name:        profileFileName,
			ContentType: profileFileContentType,
			Data:        image,
		}},
		Components:      profileComponents(targetUserID),
		AllowedMentions: &responses.AllowedMentions{},
	}, nil
}

func (m Module) lookupProfileUser(ctx context.Context, interaction interactions.Interaction, targetUserID string) (ports.DiscordUserInfo, error) {
	if m.discord != nil {
		info, err := m.discord.UserInfo(ctx, interaction.Actor.GuildID, targetUserID)
		if err == nil && strings.TrimSpace(info.ID) == "" {
			info.ID = targetUserID
		}
		if err == nil {
			return info, nil
		}
		if targetUserID != interaction.Actor.UserID {
			return ports.DiscordUserInfo{}, err
		}
	}
	return ports.DiscordUserInfo{
		ID:        interaction.Actor.UserID,
		Username:  interaction.Actor.Username,
		AvatarURL: interaction.Actor.AvatarURL,
	}, nil
}

func legacyProfileDisplayName(info ports.DiscordUserInfo, fallbackID string) string {
	name := strings.TrimSpace(info.Nickname)
	if name == "" {
		name = strings.TrimSpace(info.Username)
	}
	if name == "" {
		name = strings.TrimSpace(fallbackID)
	}
	discriminator := strings.TrimSpace(info.Discriminator)
	if discriminator != "" {
		return name + " #" + discriminator
	}
	return name
}

func profileGuildName(info ports.DiscordGuildInfo, fallbackID string) string {
	if strings.TrimSpace(info.Name) != "" {
		return info.Name
	}
	return fallbackID
}

func fetchProfileAvatar(ctx context.Context, avatarURL string) []byte {
	avatarURL = legacyCoinRankPNGURL(avatarURL)
	if avatarURL == "" || !(strings.HasPrefix(avatarURL, "http://") || strings.HasPrefix(avatarURL, "https://")) {
		return nil
	}
	reqCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(reqCtx, http.MethodGet, avatarURL, nil)
	if err != nil {
		return nil
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return nil
	}
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, io.LimitReader(response.Body, 2<<20))
	return buf.Bytes()
}

func profileComponents(userID string) []responses.ComponentRow {
	return []responses.ComponentRow{{
		Components: []responses.Component{{
			Type:     responses.ComponentTypeButton,
			CustomID: strings.TrimSpace(userID) + "my-profile",
			Label:    profileRefreshLabel,
			Emoji:    profileRefreshEmoji,
			Style:    responses.ButtonStyleSuccess,
		}},
	}}
}

func profileMissingMemberMessage() responses.Message {
	return responses.Message{
		Ephemeral: true,
		Embeds: []responses.Embed{{
			Title: profileMissingMemberTitle,
			Color: coinRankMissingUserColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func clockOrSystem(clock ports.Clock) ports.Clock {
	if clock != nil {
		return clock
	}
	return ports.SystemClock{}
}
