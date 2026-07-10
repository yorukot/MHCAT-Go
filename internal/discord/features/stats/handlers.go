package stats

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	corestats "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/stats"
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

func (m Module) CreateHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.createService.Repository == nil || m.createService.Channels == nil || m.createService.GuildStats == nil {
			return domain.ErrInvalidStatsConfigRequest
		}
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.FollowUp(ctx, statsErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		loadingMessageID := ""
		_, err := m.createService.Create(ctx, corestats.CreateRequest{
			GuildID:     interaction.Actor.GuildID,
			ChannelType: firstStatsOption(interaction, statsOptionChannelType),
			Option:      firstStatsOption(interaction, statsOptionStat),
			BotUserID:   statsBotUserID(interaction.ApplicationID, m.botUserID),
			BeforeBaseCreate: func(ctx context.Context) error {
				messageID, err := responder.CreateFollowUp(ctx, statsCreateLoadingMessage())
				if err == nil {
					loadingMessageID = messageID
				}
				return err
			},
		})
		if err != nil {
			return responder.FollowUp(ctx, statsCreateErrorMessage(err))
		}
		if loadingMessageID != "" {
			if err := responder.EditFollowUp(ctx, loadingMessageID, statsCreateSuccessMessage()); err != nil {
				return err
			}
		} else {
			if err := responder.FollowUp(ctx, statsCreateSuccessMessage()); err != nil {
				return err
			}
		}
		return m.track(ctx, interaction, StatsCreateCommandName, "stats-create")
	}
}

func (m Module) RoleHandler() interactions.Handler {
	return func(ctx context.Context, interaction interactions.Interaction, responder responses.Responder) error {
		if m.roleService.StatsRepository == nil || m.roleService.RoleRepository == nil || m.roleService.Channels == nil || m.roleService.Roles == nil {
			return domain.ErrInvalidStatsConfigRequest
		}
		if err := responder.Defer(ctx, responses.DeferOptions{}); err != nil {
			return err
		}
		loadingMessageID, err := responder.CreateFollowUp(ctx, statsCreateLoadingMessage())
		if err != nil {
			return err
		}
		if !interaction.Actor.HasPermission(permissionManageMessages) {
			return responder.EditFollowUp(ctx, loadingMessageID, statsErrorMessage("你需要有`訊息管理`才能使用此指令"))
		}
		config, err := m.roleService.Create(ctx, corestats.RoleCreateRequest{
			GuildID:     interaction.Actor.GuildID,
			ChannelType: firstStatsOption(interaction, statsOptionChannelType),
			RoleID:      firstStatsOption(interaction, statsOptionRole),
			BotUserID:   statsBotUserID(interaction.ApplicationID, m.botUserID),
		})
		if err != nil {
			return responder.EditFollowUp(ctx, loadingMessageID, statsRoleErrorMessage(err))
		}
		if err := responder.EditFollowUp(ctx, loadingMessageID, statsRoleSuccessMessage(config.ChannelID)); err != nil {
			return err
		}
		return m.track(ctx, interaction, StatsRoleCommandName, "stats-role-count")
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

func statsBotUserID(applicationID string, fallbackID string) string {
	if applicationID = strings.TrimSpace(applicationID); applicationID != "" {
		return applicationID
	}
	return strings.TrimSpace(fallbackID)
}

func statsRoleSuccessMessage(channelID string) responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "統計特定身分組成功創建",
			Description: "已成功為您創建統計特定身分組\n頻道:<#" + channelID + "> 名字可以更改喔，不要動到數字就好awa",
			Color:       statsSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func statsRoleErrorMessage(err error) responses.Message {
	switch {
	case errors.Is(err, domain.ErrInvalidStatsChannelType):
		return statsErrorMessage("你沒有進行設置要文字頻道還是語音頻道!或是你打錯了!")
	case errors.Is(err, ports.ErrStatsConfigMissing):
		return statsErrorMessage("你還沒創建過統計頻道，請先使用`/統計系統創建`")
	default:
		return statsErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
	}
}

func statsCreateSuccessMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:greentick:980496858445135893> | 成功創建!頻道(不要動到數字就沒問題)跟類別的名稱都能自行更改喔!",
			Color: statsSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func statsCreateLoadingMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title: "<a:lodding:980493229592043581> | 正在進行設置中!",
			Color: statsSuccessColor,
		}},
		AllowedMentions: &responses.AllowedMentions{},
	}
}

func statsCreateErrorMessage(err error) responses.Message {
	switch {
	case errors.Is(err, domain.ErrInvalidStatsChannelType):
		return statsErrorMessage("你沒有進行設置要文字頻道還是語音頻道!或是你打錯了!")
	case errors.Is(err, domain.ErrStatsOptionRequired):
		return statsErrorMessage("由於你已經創建過了，所以你必須說明你要創建的統計名稱，或是刪除現有的統計資料(使用統計資料刪除)!")
	case errors.Is(err, domain.ErrStatsChannelAlreadyExists):
		return statsErrorMessage("這個統計你已經創建過了!")
	case errors.Is(err, domain.ErrInvalidStatsOption):
		return statsErrorMessage("沒有這項統計可以創建欸QQ")
	default:
		return statsErrorMessage("很抱歉，出現了未知的錯誤，請重試!")
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

func firstStatsOption(interaction interactions.Interaction, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(interaction.Options[name]); value != "" {
			return value
		}
		if option, ok := interaction.CommandOptions[name]; ok {
			return strings.TrimSpace(option.String)
		}
	}
	return ""
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
