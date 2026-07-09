package stats

import (
	"context"
	"errors"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/interactions"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	permissionManageMessages = int64(8192)
	statsSuccessColor        = 0x57F287
	statsErrorColor          = 0xED4245
)

func (m Module) QueryHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if err := responder.Reply(ctx, legacyQueryMessage(m.color())); err != nil {
			return err
		}
		return m.track(ctx, interaction, StatsQueryCommandName, "stats-query")
	}
}

func (m Module) DeleteHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.service.Repository == nil {
			return domain.ErrInvalidStatsConfigRequest
		}
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditOriginal(ctx, statsErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		config, err := m.service.Delete(ctx, interaction.Actor.GuildID)
		if err != nil {
			if errors.Is(err, ports.ErrStatsConfigMissing) {
				return responder.EditOriginal(ctx, statsErrorMessage("你還沒有創建過統計數據，是要刪除甚麼啦!"))
			}
			return responder.EditOriginal(ctx, statsErrorMessage("很抱歉，出現了未知的錯誤，請重試!"))
		}
		if err := responder.EditOriginal(ctx, statsDeleteSuccessMessage(config.ParentID)); err != nil {
			return err
		}
		return m.track(ctx, interaction, StatsDeleteCommandName, "stats-delete")
	}
}

func legacyQueryMessage(color int) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "統計系統查詢",
			Description: `
        我的統計系統是每**10分鐘更新一次**` + "`(因為discord api每10分鐘才能更新一次)`" + `
        輸入 /統計系統創建 [選擇要` + "`文字頻道`" + `或是` + "`語音頻道`" + `] [輸入想創建的統計名稱]
        
        **用戶查詢**
        ` + "```" + `
用戶總數 (伺服器的總人數)
使用者總數 (伺服器非機器人人數)
機器人數 (伺服器總共的機器人數量)` + "```" + `
        **伺服器頻道**
        ` + "```" + `
頻道數量 (頻道總數量)
文字頻道數量 (文字頻道總數)
語音頻道數量 (語音頻道總數)` + "```" + `
        `,
			Color: color,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func statsDeleteSuccessMessage(parentID string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:greentick:980496858445135893> | 成功刪除，該類別以下的頻道我已經管不了囉!(類別id:" + parentID + ")",
			Color: statsSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func statsErrorMessage(content string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:Discord_AnimatedNo:1015989839809757295> | " + content,
			Color: statsErrorColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func (m Module) track(ctx context.Context, interaction interactions.Interaction, commandName string, feature string) error {
	if m.usage == nil {
		return nil
	}
	return m.usage.TrackCommand(ctx, ports.UsageEvent{
		CommandName: commandName,
		UserID:      interaction.Actor.UserID,
		GuildID:     interaction.Actor.GuildID,
		Feature:     feature,
	})
}
