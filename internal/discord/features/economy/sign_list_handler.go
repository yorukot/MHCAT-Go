package economy

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	signInListTitle          = "簽到人數資訊"
	signInListFileName       = "discord.txt"
	signInListFileType       = "text/plain; charset=utf-8"
	disappearedUserLabel     = "使用者已消失!"
	signInListOverflowNotice = "**由於人數過多，無法顯示所有成員名稱!\n請使用`.txt`檔案觀看**"
)

func (m Module) SignInListHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		result, err := m.signInList.List(ctx, interaction.Actor.GuildID, interaction.Actor.UserID, m.now())
		if err != nil {
			return err
		}
		if err := responder.EditOriginal(ctx, m.signInListMessage(ctx, interaction.Actor.GuildID, result)); err != nil {
			return err
		}
		return m.trackCommand(ctx, interaction, SignInListCommandName)
	}
}

func (m Module) signInListMessage(ctx context.Context, guildID string, result domain.SignInListResult) responses.Message {
	displayNames := make([]string, 0, len(result.Entries))
	fileLines := make([]string, 0, len(result.Entries))
	actorSigned := false
	for _, entry := range result.Entries {
		displayName := m.signInListDisplayName(ctx, guildID, entry.UserID)
		displayNames = append(displayNames, displayName)
		if entry.UserID == result.ActorUserID {
			actorSigned = true
		}
		line := fmt.Sprintf("%s(id:%s)", displayName, entry.UserID)
		if entry.ShowSignedAt {
			line += "簽到時間:" + legacySignListTime(entry.SignedAtUnix)
		}
		fileLines = append(fileLines, line)
	}
	namesText := "┃ " + strings.Join(displayNames, " ┃ ") + "┃"
	if len(displayNames) >= 100 {
		namesText = signInListOverflowNotice
	}
	description := fmt.Sprintf("<:list:992002476360343602>**目前共有**`%d`**人已經簽到**\n<:star:987020551698649138>**您是否有簽到:**%s\n\n%s", len(displayNames), legacySignInListBool(actorSigned), namesText)
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       signInListTitle,
			Description: description,
			Color:       m.randomColor(),
		}},
		Files: []responses.File{{
			Name:        signInListFileName,
			ContentType: signInListFileType,
			Data:        []byte(strings.Join(fileLines, "\n")),
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func (m Module) signInListDisplayName(ctx context.Context, guildID string, userID string) string {
	if m.discord == nil {
		return disappearedUserLabel
	}
	info, err := m.discord.UserInfo(ctx, guildID, userID)
	if err != nil || strings.TrimSpace(info.Username) == "" {
		return disappearedUserLabel
	}
	username := strings.TrimSpace(info.Username)
	discriminator := strings.TrimSpace(info.Discriminator)
	if discriminator == "" {
		return username
	}
	return username + "#" + discriminator
}

func legacySignInListBool(value bool) string {
	if value {
		return "`有`"
	}
	return "`沒有`"
}

func legacySignListTime(unixSeconds int64) string {
	location, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		location = time.FixedZone("Asia/Taipei", 8*60*60)
	}
	return time.Unix(unixSeconds, 0).In(location).Format("2006/01/02\u200915:04:05 [台北標準時間]")
}

func (m Module) randomColor() int {
	if m.color == nil {
		return coinQuerySuccessColor
	}
	return m.color()
}
