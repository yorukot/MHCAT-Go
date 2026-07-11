package utility

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	legacyInfoBotRefreshCustomID = "botinfoupdate"
	legacyInfoShardCustomID      = "shardinfoupdate"
)

var legacyLanguageFlag = map[string]string{
	"id":    "🇮🇩",
	"da":    "🇹🇦",
	"de":    "🇩🇪",
	"en-GB": "🇬🇧",
	"en-US": "🇺🇸",
	"es-ES": "🇪🇸",
	"fr":    "🇫🇷",
	"hr":    "🇭🇷",
	"it":    "🇮🇹",
	"lt":    "🇱🇹",
	"hu":    "🇭🇺",
	"nl":    "🇳🇱",
	"no":    "🇳🇴",
	"pl":    "🇵🇱",
	"pt-BR": "🇧🇷",
	"ro":    "🇷🇴",
	"fi":    "🇫🇮",
	"sv-SE": "🇸🇪",
	"vi":    "🇻🇮",
	"tr":    "🇹🇷",
	"cs":    "🇨🇿",
	"el":    "🇸🇻",
	"bg":    "🇧🇬",
	"ru":    "🇷🇺",
	"uk":    "🇺🇦",
	"hi":    "🇮🇳",
	"th":    "🇹🇭",
	"zh-CN": "🇨🇳",
	"ja":    "🇯🇵",
	"zh-TW": "🇹🇼",
	"ko":    "🇰🇷",
}

func legacyInfoBotMessage(info ports.BotInfo) responses.Message {
	return legacyInfoBotMessageWithShardLabel(info, "分片數量", false)
}

func legacyInfoBotRefreshMessage(info ports.BotInfo) responses.Message {
	return legacyInfoBotMessageWithShardLabel(info, "集群數量", true)
}

func legacyInfoBotMessageWithShardLabel(info ports.BotInfo, shardLabel string, refresh bool) responses.Message {
	if info.Name == "" {
		info.Name = "MHCAT"
	}
	if info.ShardCount == 0 {
		info.ShardCount = 1
	}
	now := time.Now()
	bootUnix := legacyBootUnix(now, info.Uptime)
	cpuModel := info.CPUModel
	if cpuModel == "" {
		cpuModel = "unknown"
	}
	memoryUsed := info.MemoryUsedMB
	memoryTotal := info.MemoryTotalMB
	memoryPercent := 0.0
	if memoryTotal > 0 {
		memoryPercent = (float64(memoryUsed) / float64(memoryTotal)) * 100
	}
	serverFieldName := "<:server:986064124209418251> 總伺服器:"
	if refresh {
		serverFieldName = ""
	}
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:     "<a:mhcat:996759164875440219> MHCAT目前系統使用量:",
			Color:     legacyUtilityRandomColor(),
			Timestamp: now,
			Fields: []responses.EmbedField{
				{Name: "<:cpu:986062422383161424> CPU型號:\n", Value: fmt.Sprintf("`%s`", cpuModel), Inline: false},
				{Name: "<:cpu:987630931932229632> CPU使用量:\n", Value: fmt.Sprintf("`%.2f`**%%**", info.CPUUsagePercent), Inline: true},
				{Name: fmt.Sprintf("<:vagueness:999527612634374184> %s:\n", shardLabel), Value: fmt.Sprintf("`%d` **個**", info.ShardCount), Inline: true},
				{Name: "<:rammemory:986062763598155797> RAM使用量:", Value: fmt.Sprintf("`%d\\%d` **MB**`(%.2f%%)`", memoryUsed, memoryTotal, memoryPercent), Inline: true},
				{Name: "<:chronometer:986065703369080884> 開機時間:", Value: fmt.Sprintf("**<t:%d:R>**", bootUnix), Inline: true},
				{Name: serverFieldName, Value: fmt.Sprintf("`%d`", info.GuildCount), Inline: true},
				{Name: "<:user:986064391139115028> 總使用者:", Value: fmt.Sprintf("`%d`", info.UserCount), Inline: true},
			},
		}},
		Components: []responses.ComponentRow{{
			Components: []responses.Component{{
				Type:     responses.ComponentTypeButton,
				CustomID: legacyInfoBotRefreshCustomID,
				Label:    "更新",
				Emoji:    "<:update:1020532095212335235>",
				Style:    responses.ButtonStyleSuccess,
			}},
		}},
	}
}

func legacyInfoErrorMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<a:Discord_AnimatedNo:1015989839809757295> | 錯誤",
			Description: "無法獲取機器人資訊，請稍後再試。",
			Color:       0xEA0000,
		}},
		Ephemeral: true,
	}
}

func legacyInfoLookupErrorMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:       "<a:Discord_AnimatedNo:1015989839809757295> | 錯誤",
			Description: "無法獲取資訊，請稍後再試。",
			Color:       0xEA0000,
		}},
	}
}

func legacyInfoShardMessage() responses.Message {
	return responses.Message{
		Embeds: []responses.Embed{{
			Title:     "<:vagueness:999527612634374184> 以下是每個分片的資訊!!",
			Color:     legacyUtilityRandomColor(),
			Timestamp: time.Now(),
		}},
		Components: []responses.ComponentRow{{
			Components: []responses.Component{{
				Type:     responses.ComponentTypeButton,
				CustomID: legacyInfoShardCustomID,
				Label:    "更新",
				Emoji:    "<:update:1020532095212335235>",
				Style:    responses.ButtonStyleSuccess,
			}},
		}},
	}
}

func legacyInfoShardRefreshMessage(info ports.BotInfo) responses.Message {
	msg := legacyInfoShardMessage()
	msg.Embeds[0].Fields = []responses.EmbedField{legacyShardField(info)}
	return msg
}

func legacyShardField(info ports.BotInfo) responses.EmbedField {
	memoryUsed := info.MemoryUsedMB
	memoryTotal := info.MemoryTotalMB
	return responses.EmbedField{
		Name: fmt.Sprintf("<:server:986064124209418251> 分片ID: %d", info.ShardID),
		Value: fmt.Sprintf("```fix\n公會數量: %d\n使用者數量: %d\n記憶體: %d\\%d mb\n上線時間:%s\n延遲: %s```",
			info.GuildCount,
			info.UserCount,
			memoryUsed,
			memoryTotal,
			formatLegacyShardUptime(info.Uptime),
			formatLegacyShardPing(info.Latency),
		),
		Inline: true,
	}
}

func formatLegacyShardUptime(value time.Duration) string {
	if value < 0 {
		value = 0
	}
	totalSeconds := int64(value / time.Second)
	hours := totalSeconds / 3600
	minutes := totalSeconds % 3600 / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02dh%02dm%02ds", hours, minutes, seconds)
}

func formatLegacyShardPing(value time.Duration) string {
	return strconv.FormatInt(value.Milliseconds(), 10)
}

func legacyInfoUserMessage(user ports.DiscordUserInfo) responses.Message {
	name := user.Username
	if name == "" {
		name = "使用者"
	}
	embed := responses.Embed{
		Title: fmt.Sprintf("<:info:985946738403737620> 以下是%s的資料", name),
		Color: legacyUtilityRandomColor(),
		Fields: []responses.EmbedField{
			{Name: "<:id:1010884394791207003> **使用者ID:**", Value: legacyCode(user.ID)},
			{Name: "<:page:992009288232996945> **創建時間:**", Value: legacyTime(user.CreatedAt)},
			{Name: "<:joins:956444030487642112> **加入時間:**", Value: legacyTime(user.JoinedAt)},
		},
	}
	if user.AvatarURL != "" {
		embed.Thumbnail = &responses.EmbedImage{URL: user.AvatarURL}
	}
	return responses.Message{Embeds: []responses.Embed{embed}}
}

func legacyInfoGuildMessage(guild ports.DiscordGuildInfo) responses.Message {
	name := guild.Name
	if name == "" {
		name = "此伺服器"
	}
	embed := responses.Embed{
		Title: fmt.Sprintf("以下是%s的資料", name),
		Color: legacyUtilityRandomColor(),
		Fields: []responses.EmbedField{
			{Name: "<:id:1010884394791207003> **伺服器ID:**", Value: legacyCode(guild.ID), Inline: true},
			{Name: "<:Discord_Members:1085959207725043812> **成員數量:**", Value: fmt.Sprintf("`%d`個", guild.MemberCount), Inline: true},
			{Name: "<a:BoosterBadgesRoll:1085958739313573980> **加成狀態:**", Value: fmt.Sprintf("**加成數:**`%d`\n**加成等級:**`%d`", guild.PremiumSubscriptionCount, guild.PremiumTier), Inline: true},
			{Name: "<:chronometer:986065703369080884> **創建時間:**", Value: legacyTimeWithRelative(guild.CreatedAt), Inline: true},
			{Name: "<:Guild_owner_dark_theme:1085959589071175712> **擁有者:**", Value: legacyMention(guild.OwnerID), Inline: true},
			{Name: "🤔 **Emoji數量:**", Value: fmt.Sprintf("`%d`個", guild.EmojiCount), Inline: true},
			{Name: "<:google:986870850391277609> **伺服器語言:**", Value: legacyLocale(guild.PreferredLocale), Inline: true},
			{Name: "<:tickmark:985949769224556614> **伺服器驗證等級:**", Value: legacyVerificationLevel(guild.VerificationLevel), Inline: true},
		},
	}
	if guild.IconURL != "" {
		embed.Thumbnail = &responses.EmbedImage{URL: guild.IconURL}
	}
	if guild.BannerURL != "" {
		embed.Image = &responses.EmbedImage{URL: guild.BannerURL}
	}
	return responses.Message{Embeds: []responses.Embed{embed}}
}

func legacyCode(value string) string {
	if value == "" {
		return "`未知`"
	}
	return fmt.Sprintf("`%s`", value)
}

func legacyTime(value time.Time) string {
	if value.IsZero() {
		return "`未知`"
	}
	return fmt.Sprintf("<t:%d>", legacyRoundedUnix(value))
}

func legacyTimeWithRelative(value time.Time) string {
	if value.IsZero() {
		return "`未知`"
	}
	seconds := legacyRoundedUnix(value)
	return fmt.Sprintf("<t:%d> (<t:%d:R>)", seconds, seconds)
}

func legacyMention(userID string) string {
	if userID == "" {
		return "`未知`"
	}
	return fmt.Sprintf("<@%s>", userID)
}

func legacyLocale(locale string) string {
	flag := legacyLanguageFlag[locale]
	if flag == "" {
		flag = "undefined"
	}
	return fmt.Sprintf("%s`(%s)`", flag, locale)
}

func legacyRoundedUnix(value time.Time) int64 {
	return int64(math.Floor(float64(value.UnixNano())/float64(time.Second) + 0.5))
}

func legacyBootUnix(now time.Time, uptime time.Duration) int64 {
	return int64(math.Floor(float64(now.UnixNano())/float64(time.Second) - uptime.Seconds() + 0.5))
}

func legacyVerificationLevel(level int) string {
	description := "此伺服器無任何驗證機制"
	switch level {
	case 1:
		description = "需通過電子郵件認證"
	case 2:
		description = "須通過電子郵件認證並成員dc成員5分鐘"
	case 3:
		description = "必須成為此伺服器成員10分鐘"
	case 4:
		description = "必須經過手機認證"
	}
	return fmt.Sprintf("`%d`**(%s)**", level, description)
}
